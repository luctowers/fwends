def pytest_addoption(parser):
  parser.addoption("--backend", default="http://localhost:8080", help="The backend endpoint to use")

def pytest_generate_tests(metafunc):
  backend_value = metafunc.config.option.backend
  if "backend" in metafunc.fixturenames and backend_value is not None:
    metafunc.parametrize("backend", [backend_value])
