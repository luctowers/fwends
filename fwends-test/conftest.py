import time
import requests

def pytest_addoption(parser):
  parser.addoption("--backend", default="http://localhost:8080", help="The backend endpoint to use")
  parser.addoption("--health-check-enable", action="store_true", default=False, help="Whether to wait for health check")
  parser.addoption("--health-check-timeout",  type=float, default=0, help="Max time to wait for health check")
  parser.addoption("--health-check-delay",  type=float, default=1, help="Min retry time for health check")

healthy = False
def wait_for_health_check(backend, timeout, delay):
  global healthy
  if healthy:
    return
  start = time.time()
  while True:
    try:
      response = requests.get(
        backend + "/api/health"
      )
      assert response.status_code == 200
      response_data = response.json()
      #if all services are healthy
      if all(response_data["services"].values()):
        healthy = True
        break
      else:
        raise ValueError("Not all services are healthy")
    except Exception as e:
      print(e)
      if timeout == 0:
        time.sleep(delay)
      else:
        elapsed = time.time() - start
        if elapsed + delay >= timeout:
          raise e
        else:
          time.sleep(delay)

def pytest_generate_tests(metafunc): 
  if "backend" in metafunc.fixturenames:
    backend = metafunc.config.getoption("--backend")
    health_check_enable = metafunc.config.getoption("--health-check-enable")
    health_check_timeout = metafunc.config.getoption("--health-check-timeout")
    health_check_delay = metafunc.config.getoption("--health-check-delay")
    if health_check_enable:
      wait_for_health_check(backend, health_check_timeout, health_check_delay)
    metafunc.parametrize("backend", [metafunc.config.getoption("backend")])
