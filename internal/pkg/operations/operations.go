// Package operations provides a centralized nexus for processing requests
// on Discord API operations.
package operations

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// RequestType enumerations.
const (
	CreateRole RequestType = iota
)

// RequestType string representations.
const (
	CreateRoleString = "CreateRole"
	UnknownString    = "unknown"
)

const (
	roleHoist             = true
	roleMention           = true
	guildMembersPageLimit = 1000

	apiErrorCodeMaxRoles = 30005
)

// Request is an operations request to be processed.
type Request struct {
	Type       RequestType
	CreateRole *CreateRoleRequest
}

// RequestType represents a type of operations request.
type RequestType int

// String returns the string representation for the given RequestType.
func (rt RequestType) String() string {
	switch rt {
	case CreateRole:
		return CreateRoleString
	default:
		return UnknownString
	}
}

// CreateRoleRequest is a request to create a new role.
type CreateRoleRequest struct {
	Guild     *discordgo.Guild
	RoleName  string
	RoleColor int
}

// ResultChannel is a channel the result from an operation is sent to.
type ResultChannel chan interface{}

// NewResultChannel returns a new, buffered channel to send an operation result
// to.
func NewResultChannel() ResultChannel {
	return make(ResultChannel, 1)
}

// Nexus is a centralized construct to process operation requests by
// de-duplicating identical simultaneous requests and providing the result to
// all of the callers.
type Nexus struct {
	Session *discordgo.Session

	mutex          *sync.Mutex
	resultChannels map[keyHash][]ResultChannel
}

type keyHash uint32

// NewNexus returns a new *Nexus ready to process requests.
func NewNexus(session *discordgo.Session) *Nexus {
	return &Nexus{
		Session:        session,
		mutex:          &sync.Mutex{},
		resultChannels: make(map[keyHash][]ResultChannel),
	}
}

// Process will process the provided request and send back the result to the
// provided ResultChannel. The caller should type check the result it receives
// to determine if an error was sent or the result is of the type it expects.
func (nexus *Nexus) Process(ctx context.Context, resultChannel ResultChannel, request *Request) {
	switch request.Type {
	case CreateRole:
		nexus.processCreateRole(ctx, resultChannel, request)
	default:
		resultChannel <- fmt.Errorf("%s request type not supported", request.Type)
		close(resultChannel)
	}
}

func (nexus *Nexus) processCreateRole(ctx context.Context, resultChannel ResultChannel, request *Request) {
	hashFunc := fnv.New32()

	// According to documentation, this Write will never return an error
	_, _ = hashFunc.Write([]byte(fmt.Sprintf(
		"%s/%s/%s",
		request.Type,
		request.CreateRole.Guild.ID,
		request.CreateRole.RoleName,
	)))

	key := keyHash(hashFunc.Sum32())

	nexus.mutex.Lock()

	_, found := nexus.resultChannels[key]
	if found {
		nexus.resultChannels[key] = append(nexus.resultChannels[key], resultChannel)
		nexus.mutex.Unlock()

		return
	}

	nexus.resultChannels[key] = []ResultChannel{resultChannel}
	nexus.mutex.Unlock()

	role, err := createRole(
		ctx,
		nexus.Session,
		request.CreateRole.Guild,
		request.CreateRole.RoleName,
		request.CreateRole.RoleColor,
	)
	if err != nil {
		nexus.sendResult(key, err)
		return
	}

	nexus.sendResult(key, role)
}

func (nexus *Nexus) sendResult(key keyHash, result interface{}) {
	nexus.mutex.Lock()
	defer nexus.mutex.Unlock()

	defer delete(nexus.resultChannels, key)

	for _, resultChannel := range nexus.resultChannels[key] {
		resultChannel <- result
		close(resultChannel)
	}
}

// LookupGuild returns a *discordgo.Guild from the session's internal state
// cache. If the guild is not found in the state cache, LookupGuild will query
// the Discord API for the guild and add it to the state cache before returning
// it.
func LookupGuild(ctx context.Context, session *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	guild, err := session.State.Guild(guildID)
	if err != nil {
		guild, err = updateStateGuilds(ctx, session, guildID)
		if err != nil {
			return nil, fmt.Errorf("unable to query guild: %w", err)
		}
	}

	return guild, nil
}

// AddRoleToMember adds the role associated with the provided roleID to the
// user associated with the provided userID, in the guild associated with the
// provided guildID.
func AddRoleToMember(ctx context.Context, session *discordgo.Session, guildID, userID, roleID string) error {
	err := session.GuildMemberRoleAddWithContext(ctx, guildID, userID, roleID)
	if err != nil {
		return fmt.Errorf("unable to add ephemeral role: %w", err)
	}

	return nil
}

// RemoveRoleFromMember removes the role associated with the provided roleID
// from the user associated with the provided userID, in the guild associated
// with the provided guildID.
func RemoveRoleFromMember(ctx context.Context, session *discordgo.Session, guildID, userID, roleID string) error {
	err := session.GuildMemberRoleRemoveWithContext(ctx, guildID, userID, roleID)
	if err != nil {
		return fmt.Errorf("unable to remove ephemeral role: %w", err)
	}

	return nil
}

// IsForbiddenResponse checks if the provided error wraps *discordgo.RESTError.
// If it does, IsForbiddenResponse returns true if the response code is equal
// to http.StatusForbidden.
func IsForbiddenResponse(err error) bool {
	var restErr *discordgo.RESTError

	if errors.As(err, &restErr) {
		if restErr.Response.StatusCode == http.StatusForbidden {
			return true
		}
	}

	return false
}

// IsMaxGuildsResponse checks if the provided error wraps *discordgo.RESTError.
// If it does, IsMaxGuildsResponse returns true if the response code is equal
// to http.StatusBadRequest and the error code is 30005.
func IsMaxGuildsResponse(err error) bool {
	var restErr *discordgo.RESTError

	if errors.As(err, &restErr) {
		if restErr.Response.StatusCode == http.StatusBadRequest {
			return restErr.Message.Code == apiErrorCodeMaxRoles
		}
	}

	return false
}

// BotHasChannelPermission checks if the bot has view permissions for the
// channel. If the bot does have the view permission, BotHasChannelPermission
// returns nil.
func BotHasChannelPermission(ctx context.Context, session *discordgo.Session, channel *discordgo.Channel) error {
	permissions, err := session.UserChannelPermissions(session.State.User.ID, channel.ID)
	if err != nil {
		return fmt.Errorf("unable to determine channel permissions: %w", err)
	}

	if permissions&discordgo.PermissionViewChannel != discordgo.PermissionViewChannel {
		return fmt.Errorf("insufficient channel permissions: channel: %s", channel.Name)
	}

	return nil
}

func updateStateGuilds(ctx context.Context, session *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	guild, err := session.GuildWithContext(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("error senging guild query request: %w", err)
	}

	roles, err := session.GuildRolesWithContext(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("unable to query guild channels: %w", err)
	}

	channels, err := session.GuildChannelsWithContext(ctx, guildID)
	if err != nil {
		return nil, fmt.Errorf("unable to query guild channels: %w", err)
	}

	members, err := recursiveGuildMembersWithContext(ctx, session, guildID, "", guildMembersPageLimit)
	if err != nil {
		return nil, fmt.Errorf("unable to query guild members: %w", err)
	}

	guild.Roles = roles
	guild.Channels = channels
	guild.Members = members
	guild.MemberCount = len(members)

	err = session.State.GuildAdd(guild)
	if err != nil {
		return nil, fmt.Errorf("unable to add guild to state cache: %w", err)
	}

	return guild, nil
}

func createRole(
	ctx context.Context,
	session *discordgo.Session,
	guild *discordgo.Guild,
	roleName string,
	roleColor int,
) (*discordgo.Role, error) {
	role, err := session.GuildRoleCreateWithContext(ctx, guild.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to create ephemeral role: %w", err)
	}

	role, err = session.GuildRoleEditWithContext(
		ctx,
		guild.ID, role.ID,
		roleName, roleColor,
		roleHoist, role.Permissions, roleMention,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to edit ephemeral role: %w", err)
	}

	err = session.State.RoleAdd(guild.ID, role)
	if err != nil {
		return nil, fmt.Errorf("unable to add ephemeral role to state cache: %w", err)
	}

	return role, nil
}

func recursiveGuildMembersWithContext(
	ctx context.Context,
	session *discordgo.Session,
	guildID, after string,
	limit int,
) ([]*discordgo.Member, error) {
	guildMembers, err := session.GuildMembersWithContext(ctx, guildID, after, limit)
	if err != nil {
		return nil, fmt.Errorf("error sending recursive guild members request: %w", err)
	}

	if len(guildMembers) < guildMembersPageLimit {
		return guildMembers, nil
	}

	nextGuildMembers, err := recursiveGuildMembersWithContext(
		ctx,
		session,
		guildID,
		guildMembers[len(guildMembers)-1].User.ID,
		guildMembersPageLimit,
	)
	if err != nil {
		return nil, err
	}

	return append(guildMembers, nextGuildMembers...), nil
}
