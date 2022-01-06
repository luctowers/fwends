import pytest
import requests
import kubernetes.client
from util import (
	get_replica_count,
	get_desired_replica_count,
	set_replica_count,
	get_replica_count_stateful,
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
		"s3": True,
	}})


@pytest.mark.failure_test
def test_health_check_redis(backend, kubernetes_client, namespace):
	"""Test redis service is marked as unhealthy when it is down."""

	appsv1 = kubernetes.client.AppsV1Api(kubernetes_client)

	# tell redis to stop
	restore_replica_count = get_desired_replica_count(
		appsv1, namespace, "fwends-redis"
	)
	set_replica_count(appsv1, namespace, "fwends-redis", 0)

	# wait for postgres to stop
	def assert_redis_down():
		assert get_replica_count(appsv1, namespace, "fwends-redis") == 0
	retry_assert(assert_redis_down, 10)

	try:
		# assert redis unhealthy
		assert_health_check(backend, {"services": {
			"postgres": True,
			"redis": False,
			"s3": True,
		}})
	finally:
		# restore redis
		set_replica_count(
			appsv1, namespace, "fwends-redis", restore_replica_count
		)

	# wait for all services to be healthy again
	def assert_healthy():
		assert_health_check(backend, {"services": {
			"postgres": True,
			"redis": True,
			"s3": True,
		}})
	retry_assert(assert_healthy, 10)


@pytest.mark.failure_test
def test_health_check_postgres(backend, kubernetes_client, namespace):
	"""Test postgres service is marked as unhealthy when it is down."""

	appsv1 = kubernetes.client.AppsV1Api(kubernetes_client)

	# tell postgres to stop
	restore_replica_count = get_desired_replica_count_stateful(
		appsv1, namespace, "fwends-postgres"
	)
	set_replica_count_stateful(appsv1, namespace, "fwends-postgres", 0)

	# wait for postgres to stop
	def assert_redis_down():
		assert get_replica_count_stateful(appsv1, namespace, "fwends-postgres") == 0
	retry_assert(assert_redis_down, 10)

	try:
		# assert postgres unhealthy
		assert_health_check(backend, {"services": {
			"postgres": False,
			"redis": True,
			"s3": True,
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
			"s3": True,
		}})
	retry_assert(assert_healthy, 10)


@pytest.mark.failure_test
def test_health_check_s3(backend, kubernetes_client, namespace):
	"""Test postgres service is marked as unhealthy when it is down."""

	appsv1 = kubernetes.client.AppsV1Api(kubernetes_client)

	# tell minio to stop
	restore_replica_count = get_desired_replica_count_stateful(
		appsv1, namespace, "fwends-minio"
	)
	set_replica_count_stateful(appsv1, namespace, "fwends-minio", 0)

	# wait for minio to stop
	def assert_minio_down():
		assert get_replica_count_stateful(appsv1, namespace, "fwends-minio") == 0
	retry_assert(assert_minio_down, 10)

	try:
		# assert minio unhealthy
		assert_health_check(backend, {"services": {
			"postgres": True,
			"redis": True,
			"s3": False,
		}})
	finally:
		# restore minio
		set_replica_count_stateful(
			appsv1, namespace, "fwends-minio", restore_replica_count
		)

	# wait for all services to be healthy again
	def assert_healthy():
		assert_health_check(backend, {"services": {
			"postgres": True,
			"redis": True,
			"s3": True,
		}})
	retry_assert(assert_healthy, 10)
