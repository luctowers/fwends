import pytest
import kubernetes
from util import wait_for_health_check


def pytest_configure(config):
	config.addinivalue_line(
		"markers", "failure_test: marks as a failure test that may be skipped"
	)


def pytest_addoption(parser):
	parser.addoption(
		"--namespace",
		default="default",
		help="The kubernetes namespace to use"
	)
	parser.addoption(
		"--backend",
		default="http://localhost:8080/api",
		help="The backend endpoint to use"
	)
	parser.addoption(
		"--media",
		default="http://localhost:8080/media",
		help="The media endpoint to use"
	)
	parser.addoption(
		"--kube-proxy",
		default="http://localhost:8081",
		help="The kubernetes api proxy to use"
	)
	parser.addoption(
		"--health-check-enable",
		action="store_true",
		default=False,
		help="Whether to wait for health check"
	)
	parser.addoption(
		"--health-check-timeout",
		type=float,
		default=0,
		help="Max time to wait for health check"
	)
	parser.addoption(
		"--health-check-delay",
		type=float,
		default=1,
		help="Min retry time for health check"
	)
	parser.addoption(
		"--failure-test-enable",
		action="store_true",
		default=False,
    help="Include failure tests"
	)


def pytest_runtest_setup(item):
	is_failure_test = 'failure_test' in item.keywords
	if is_failure_test and not item.config.getoption("--failure-test-enable"):
		pytest.skip("need --failure-tests option to run this test")


@pytest.fixture
def backend(request):
	backend_val = request.config.getoption("--backend")
	health_check_enable = request.config.getoption("--health-check-enable")
	if health_check_enable and not backend.healthy:
		if backend.error:
			raise AssertionError("backend is unhealthy")
		health_check_timeout = request.config.getoption("--health-check-timeout")
		health_check_delay = request.config.getoption("--health-check-delay")
		try:
			wait_for_health_check(backend_val, health_check_timeout, health_check_delay)
			backend.healthy = True
		except AssertionError:
			backend.error = True
	return backend_val
backend.healthy = False
backend.error = False

@pytest.fixture
def media(request):
	return request.config.getoption("--media")

@pytest.fixture
def namespace(request):
	return request.config.getoption("--namespace")


@pytest.fixture
def kubernetes_client(request):
	kubernetes.config.load_kube_config()
	configuration = kubernetes.client.Configuration()
	configuration.proxy = request.config.getoption("--kube-proxy")
	with kubernetes.client.ApiClient(configuration) as api_client:
		return api_client
