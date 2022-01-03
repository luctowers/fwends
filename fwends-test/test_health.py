import pytest
import requests
import kubernetes.client
from util import (
	get_desired_replica_count_stateful,
	set_replica_count_stateful,
	retry_assert
)


def assert_health_check(backend, assert_data):
	"""Assert health check response from backend is exactly assert_data."""

	response = requests.get(
		backend + "/api/health/"
	)
	assert response.status_code == 200
	response_data = response.json()
	assert response_data == assert_data


def test_health_check_healthy(backend):
	"""Test all services are healthy in the default state."""

	assert_health_check(backend, {"services": {
		"postgres": True,
		"redis": True,
	}})


@pytest.mark.failure_test
def test_health_check_redis(backend, kubernetes_client, namespace):
	"""Test redis service is marked as unhealthy when it is down."""

	appsv1 = kubernetes.client.AppsV1Api(kubernetes_client)

	# bring down redis
	restore_replica_count = get_desired_replica_count_stateful(
		appsv1, namespace, "fwends-redis"
	)
	set_replica_count_stateful(appsv1, namespace, "fwends-redis", 0)

	try:
		# assert redis unhealthy
		assert_health_check(backend, {"services": {
			"postgres": True,
			"redis": False,
		}})
	finally:
		# restore redis
		set_replica_count_stateful(
			appsv1, namespace, "fwends-redis", restore_replica_count
		)

	# wait for all services to be healthy again
	def assert_healthy():
		assert_health_check(backend, {"services": {
			"postgres": True,
			"redis": True,
		}})
	retry_assert(assert_healthy, 60)

@pytest.mark.failure_test
def test_health_check_postgres(backend, kubernetes_client, namespace):
	"""Test postgres service is marked as unhealthy when it is down."""

	appsv1 = kubernetes.client.AppsV1Api(kubernetes_client)

	# bring down postgres
	restore_replica_count = get_desired_replica_count_stateful(
		appsv1, namespace, "fwends-postgres"
	)
	set_replica_count_stateful(appsv1, namespace, "fwends-postgres", 0)

	try:
		# assert postgres unhealthy
		assert_health_check(backend, {"services": {
			"postgres": False,
			"redis": True,
		}})
	finally:
		# restore postgres
		set_replica_count_stateful(
			appsv1, namespace, "fwends-postgres", restore_replica_count
		)

	# wait for all services to be healthy again
	def assert_healthy():
		assert_health_check(backend, {"services": {
			"postgres": True,
			"redis": True,
		}})
	retry_assert(assert_healthy, 60)
