# FWENDS Integration and Failure Tests

## Prerequisites

1. Complete minimal dev setup [README.md](../README.md)
2. install python https://www.python.org/downloads/
3. install python poetry https://python-poetry.org/docs/#installation
4. install python dependencies `cd fwends-test && poetry install`

## Integration tests only

1. Run services with skaffold eg. `skaffold dev -p dev`, or `skaffold run --tail`
2. Run tests `cd fwends-test && poetry run pytest`

## Integration tests and failure tests

1. Run services with skaffold (as above)
2. Run kubernetes api proxy `kubectl proxy --port 8081`
3. Run tests with failure tests enabled  `cd fwends-test && poetry run pytest --failure-test-enable`
