import requests


def test_pack_crud(backend):
	"""Test the POST, GET, PUT and DELETE methods of packs api."""

	# create pack
	pack = {"title":"Test Title!"}
	response = requests.post(
		backend + "/api/packs/",
		json=pack
	)
	assert response.status_code == 200
	response_data = response.json()
	assert "id" in response_data
	pack_id = response_data["id"]
	assert isinstance(pack_id, str)
	assert int(pack_id) >= 0

	# get pack
	response = requests.get(
		backend + "/api/packs/" + pack_id
	)
	assert response.status_code == 200
	response_data = response.json()
	assert "title" in response_data
	assert response_data["title"] == pack["title"]

	# update pack
	updated_pack = {"title":"Updated Title!!"}
	response = requests.put(
		backend + "/api/packs/" + pack_id,
		json=updated_pack
	)
	assert response.status_code == 200

	# get pack again
	response = requests.get(
		backend + "/api/packs/" + pack_id
	)
	assert response.status_code == 200
	response_data = response.json()
	assert "title" in response_data
	assert response_data["title"] == updated_pack["title"]

	# delete pack
	response = requests.delete(
		backend + "/api/packs/" + pack_id
	)
	assert response.status_code == 200

	# get pack not found
	response = requests.get(
		backend + "/api/packs/" + pack_id
	)
	assert response.status_code == 404
