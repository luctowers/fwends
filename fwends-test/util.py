import time
import requests

def wait_for_health_check(backend, timeout, delay):
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
				wait_for_health_check.healthy = True
				break
			raise ValueError("Not all services are healthy")
		except (
			ValueError, AssertionError, requests.exceptions.RequestException
		) as err:
			print(err)
			if timeout != 0:
				elapsed = time.time() - start
				if elapsed + delay >= timeout:
					raise err
			time.sleep(delay)


def retry_assert(function, timeout=0, delay=1):
	start = time.time()
	while True:
		try:
			return function()
		except AssertionError as err:
			if timeout != 0:
				elapsed = time.time() - start
				if elapsed + delay >= timeout:
					raise err
			time.sleep(delay)


def get_desired_replica_count_stateful(appsv1, namespace, app):
	stateful_set = appsv1.list_namespaced_stateful_set(
		namespace, field_selector="metadata.name=" + app
	).items[0]
	return stateful_set.spec.replicas


def get_replica_count_stateful(appsv1, namespace, app):
	replica_set = appsv1.list_namespaced_stateful_set(
		namespace, field_selector="metadata.name=" + app
	).items[0]
	return replica_set.status.replicas


def set_replica_count_stateful(appsv1, namespace, app, count):
	body = {"spec": {"replicas": count}}
	appsv1.patch_namespaced_stateful_set_scale(app, namespace, body)

def get_desired_replica_count(appsv1, namespace, app):
	stateful_set = appsv1.list_namespaced_deployment(
		namespace, field_selector="metadata.name=" + app
	).items[0]
	return stateful_set.spec.replicas


def get_replica_count(appsv1, namespace, app):
	replica_set = appsv1.list_namespaced_deployment(
		namespace, field_selector="metadata.name=" + app
	).items[0]
	if replica_set.status.replicas is None:
		return 0
	return replica_set.status.replicas


def set_replica_count(appsv1, namespace, app, count):
	body = {"spec": {"replicas": count}}
	appsv1.patch_namespaced_deployment_scale(app, namespace, body)
