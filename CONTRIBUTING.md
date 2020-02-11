## Contributing to the project

Contributions are very welcome! Please follow the guidelines below:

* Open an issue describing the bug or enhancement
* Fork the `develop` branch and make your changes
  * Try to match current naming conventions as closely as possible
  * Try to keep changes small and incremental with appropriate new unit tests
* Create a Pull Request with your changes against the `develop` branch

This project is equipped with a full
[CI](https://en.wikipedia.org/wiki/Continuous_integration)
/
[CD](https://en.wikipedia.org/wiki/Continuous_deployment) pipeline:
 
* Linting and unit tests will be automatically run with the PR, providing
feedback if any additional changes need to be made
* Merge to `master` will automatically deploy the changes live