import hashlib
import requests


def test_pack_crud(backend):
	"""Test the POST, GET, PUT and DELETE methods of packs api."""

	# create pack
	pack = {"title":"Test Title!"}
	response = requests.post(
		backend + "/packs/",
		json=pack
	)
	assert response.status_code == 200
	response_data = response.json()
	assert "id" in response_data
	pack_id = response_data["id"]
	assert isinstance(pack_id, str)
	assert int(pack_id) >= 0

	# get pack
	response = requests.get(backend+"/packs/"+pack_id)
	assert response.status_code == 200
	response_data = response.json()
	assert "title" in response_data
	assert response_data["title"] == pack["title"]

	# verify with list pack api
	response = requests.get(backend+"/packs/")
	assert response.status_code == 200
	response_element = next((x for x in response.json() if x['id']==pack_id),None)
	assert response_element == {
		"id": pack_id,
		"title": pack["title"],
		"roleCount": 0,
		"stringCount": 0
	}

	# update pack
	updated_pack = {"title":"Updated Title!!"}
	response = requests.put(backend+"/packs/"+pack_id, json=updated_pack)
	assert response.status_code == 200

	# get pack again
	response = requests.get(backend+"/packs/"+pack_id)
	assert response.status_code == 200
	response_data = response.json()
	assert "title" in response_data
	assert response_data["title"] == updated_pack["title"]

	# verify with list pack api
	response = requests.get(backend+"/packs/")
	assert response.status_code == 200
	response_element = next((x for x in response.json() if x['id']==pack_id),None)
	assert response_element == {
		"id": pack_id,
		"title": updated_pack["title"],
		"roleCount": 0,
		"stringCount": 0
	}

	# delete pack
	response = requests.delete(backend+"/packs/"+pack_id)
	assert response.status_code == 200

	# get pack not found
	response = requests.get(backend+"/packs/"+pack_id)
	assert response.status_code == 404


def test_pack_resource_crud(backend, media):

	# create pack
	pack = {"title":"Test Title!"}
	response = requests.post(backend + "/packs/", json=pack)
	assert response.status_code == 200
	response_data = response.json()
	pack_id = response_data["id"]
	pack_api = backend+"/packs/"+pack_id
	pack_media = media+"/packs/"+pack_id

	# upload resources
	with open("./resources/bird-duck.jpg", "rb") as file:
		upload_resource(pack_api+"/bird/duck", file, "image/jpeg")
	with open("./resources/bird-duck.aac", "rb") as file:
		upload_resource(pack_api+"/bird/duck", file, "audio/aac")
	with open("./resources/bird-eagle.png", "rb") as file:
		upload_resource(pack_api+"/bird/eagle", file, "image/png")
	with open("./resources/bird-eagle.mp3", "rb") as file:
		upload_resource(pack_api+"/bird/eagle", file, "audio/mpeg")
	# omit robin image
	# with open("./resources/bird-robin.jpg", "rb") as file:
	# 	upload_resource(pack_api+"/bird/robin", file, "image/jpeg")
	with open("./resources/bird-robin.mp3", "rb") as file:
		upload_resource(pack_api+"/bird/robin", file, "audio/mpeg")
	with open("./resources/mammal-cat.jpg", "rb") as file:
		upload_resource(pack_api+"/mammal/cat", file, "image/jpeg")
	with open("./resources/mammal-cat.flac", "rb") as file:
		upload_resource(pack_api+"/mammal/cat", file, "audio/flac")
	with open("./resources/mammal-dog.webp", "rb") as file:
		upload_resource(pack_api+"/mammal/dog", file, "image/webp")
	with open("./resources/mammal-dog.m4a", "rb") as file:
		upload_resource(pack_api+"/mammal/dog", file, "audio/aac")
	with open("./resources/mammal-tiger.svg", "rb") as file:
		upload_resource(pack_api+"/mammal/tiger", file, "image/svg+xml")
	# omit tiger sound
	# with open("./resources/mammal-tiger.wav", "rb") as file:
	# 	upload_resource(pack_api+"/mammal/tiger", file, "audio/wav")

	# verify resources have been uploaded
	with open("./resources/bird-duck.jpg", "rb") as file:
		verify_resource(pack_media+"/bird/duck/image", file, "image/jpeg")
	with open("./resources/bird-duck.aac", "rb") as file:
		verify_resource(pack_media+"/bird/duck/audio", file, "audio/aac")
	with open("./resources/bird-eagle.png", "rb") as file:
		verify_resource(pack_media+"/bird/eagle/image", file, "image/png")
	with open("./resources/bird-eagle.mp3", "rb") as file:
		verify_resource(pack_media+"/bird/eagle/audio", file, "audio/mpeg")
	# omit robin image
	# with open("./resources/bird-robin.jpg", "rb") as file:
	# 	verify_resource(pack_media+"/bird/robin/image", file, "image/jpeg")
	with open("./resources/bird-robin.mp3", "rb") as file:
		verify_resource(pack_media+"/bird/robin/audio", file, "audio/mpeg")
	with open("./resources/mammal-cat.jpg", "rb") as file:
		verify_resource(pack_media+"/mammal/cat/image", file, "image/jpeg")
	with open("./resources/mammal-cat.flac", "rb") as file:
		verify_resource(pack_media+"/mammal/cat/audio", file, "audio/flac")
	with open("./resources/mammal-dog.webp", "rb") as file:
		verify_resource(pack_media+"/mammal/dog/image", file, "image/webp")
	with open("./resources/mammal-dog.m4a", "rb") as file:
		verify_resource(pack_media+"/mammal/dog/audio", file, "audio/aac")
	with open("./resources/mammal-tiger.svg", "rb") as file:
		verify_resource(pack_media+"/mammal/tiger/image", file, "image/svg+xml")
	# omit tiger sound
	# with open("./resources/mammal-tiger.wav", "rb") as file:
	# 	verify_resource(pack_media+"/mammal/tiger/audio", file, "audio/wav")

	# verify with get pack api
	response = requests.get(pack_api)
	assert response.status_code == 200
	assert response.json()["resources"] == {
		"bird": {
			"duck": {"image": True, "audio": True},
			"eagle": {"image": True, "audio": True},
			"robin": {"image": False, "audio": True},
		},
		"mammal": {
			"cat": {"image": True, "audio": True},
			"dog": {"image": True, "audio": True},
			"tiger": {"image": True, "audio": False}
		}
	}

	# verify with list pack api
	response = requests.get(backend+"/packs/")
	assert response.status_code == 200
	response_element = next((x for x in response.json() if x['id']==pack_id),None)
	assert response_element == {
		"id": pack_id,
		"title": pack["title"],
		"roleCount": 2, # mammal and bird
		"stringCount": 4 # tiger and robin, not counted as resources were omitted
	}

	# delete cat string
	response = requests.delete(pack_api+"/mammal/cat")
	assert response.status_code == 200
	# make sure tiger resources have been deleted
	verify_resource_deleted(pack_media+"/mammal/cat/image")
	verify_resource_deleted(pack_media+"/mammal/cat/audio")

	# verify with get pack api
	response = requests.get(pack_api)
	assert response.status_code == 200
	assert response.json()["resources"] == {
		"bird": {
			"duck": {"image": True, "audio": True},
			"eagle": {"image": True, "audio": True},
			"robin": {"image": False, "audio": True},
		},
		"mammal": {
			# cat deleted
			"dog": {"image": True, "audio": True},
			"tiger": {"image": True, "audio": False}
		}
	}

	# verify with list pack api
	response = requests.get(backend+"/packs/")
	assert response.status_code == 200
	response_element = next((x for x in response.json() if x['id']==pack_id),None)
	assert response_element == {
		"id": pack_id,
		"title": pack["title"],
		"roleCount": 2, # mammal and bird
		"stringCount": 3 # cat now deleted
	}

	# delete bird role
	response = requests.delete(pack_api+"/bird")
	assert response.status_code == 200
	# make sure bird resources have been deleted
	verify_resource_deleted(pack_media+"/bird/duck/image")
	verify_resource_deleted(pack_media+"/bird/duck/audio")
	verify_resource_deleted(pack_media+"/bird/eagle/image")
	verify_resource_deleted(pack_media+"/bird/eagle/audio")
	# omit robin image
	# verify_resource_deleted(pack_media+"/bird/robin/image")
	verify_resource_deleted(pack_media+"/bird/robin/audio")

	# verify others mammals are still alive
	with open("./resources/mammal-tiger.svg", "rb") as file:
		verify_resource(pack_media+"/mammal/tiger/image", file, "image/svg+xml")
	# omit tiger sound
	# with open("./resources/mammal-tiger.wav", "rb") as file:
	# 	verify_resource(pack_media+"/mammal/tiger/audio", file, "audio/wav")
	with open("./resources/mammal-dog.webp", "rb") as file:
		verify_resource(pack_media+"/mammal/dog/image", file, "image/webp")
	with open("./resources/mammal-dog.m4a", "rb") as file:
		verify_resource(pack_media+"/mammal/dog/audio", file, "audio/aac")

	# replace dog with eagle to test updating resource
	with open("./resources/bird-eagle.png", "rb") as file:
		upload_resource(pack_api+"/mammal/dog", file, "image/png")
	with open("./resources/bird-eagle.mp3", "rb") as file:
		upload_resource(pack_api+"/mammal/dog", file, "audio/mpeg")
	# verify resource has been updated
	with open("./resources/bird-eagle.png", "rb") as file:
		verify_resource(pack_media+"/mammal/dog/image", file, "image/png")
	with open("./resources/bird-eagle.mp3", "rb") as file:
		verify_resource(pack_media+"/mammal/dog/audio", file, "audio/mpeg")

	# verify with get pack api
	response = requests.get(pack_api)
	assert response.status_code == 200
	assert response.json()["resources"] == {
		# birds deleted
		"mammal": {
			"dog": {"image": True, "audio": True},
			"tiger": {"image": True, "audio": False}
		}
	}

	# verify with list pack api
	response = requests.get(backend+"/packs/")
	assert response.status_code == 200
	response_element = next((x for x in response.json() if x['id']==pack_id),None)
	assert response_element == {
		"id": pack_id,
		"title": pack["title"],
		"roleCount": 1, # only mammal is left
		"stringCount": 1 # only dog is left
	}

	# delete pack
	response = requests.delete(pack_api)
	assert response.status_code == 200
	# make sure remaining pack resources have been deleted
	verify_resource_deleted(pack_media+"/mammal/dog/image")
	verify_resource_deleted(pack_media+"/mammal/dog/audio")
	verify_resource_deleted(pack_media+"/mammal/tiger/image")
	# omit tiger sound
	# verify_resource_deleted(pack_media+"/mammal/tiger/audio")


# HELPERS

def hash_file(file):
	sha256 = hashlib.sha256()
	while True:
		data = file.read(65536)
		if not data:
			break
		sha256.update(data)
	return sha256.hexdigest()

def hash_response(res):
	sha256 = hashlib.sha256()
	for chunk in res.iter_content(65536):
		sha256.update(chunk)
	return sha256.hexdigest()

def upload_resource(url, file, content_type):
	file.seek(0)
	response = requests.put(url, headers={"Content-Type":content_type}, data=file)
	assert response.status_code == 200

def verify_resource(url, file, content_type):
	file.seek(0)
	response = requests.get(url, stream=True)
	assert response.status_code == 200
	assert response.headers["Content-Type"] == content_type
	assert hash_response(response) == hash_file(file)

def verify_resource_deleted(url):
	response = requests.get(url)
	assert response.status_code == 404
