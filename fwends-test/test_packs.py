import hashlib
import requests

def test_pack_no_resources(backend):
	pack_id = create_test_pack(backend, "Test Pack Update")
	verify_pack_title(backend, pack_id, "Test Pack Update")

	updated_pack = {"title":"New Updated Title!"}
	response = requests.put(backend+"/packs/"+pack_id, json=updated_pack)
	assert response.status_code == 200
	verify_pack_title(backend, pack_id, "New Updated Title!")

	response = requests.delete(backend+"/packs/"+pack_id)
	assert response.status_code == 200

	response = requests.get(backend+"/packs/"+pack_id)
	assert response.status_code == 404

def test_pack_hash(backend, media):
	pack_id = create_test_pack(backend, "Test Pack Hash")
	expected_hash = hashlib.sha256(b'').hexdigest()
	verify_pack_hash(backend, pack_id, expected_hash)

	populate_test_pack_resources(backend, media, pack_id)
	expected_hash = hashlib.sha256(
		'\x00bird\x01duck\x01eagle\x01robin\x00mammal\x01cat\x01dog\x01tiger'.encode('utf-8')
	).hexdigest()
	verify_pack_hash(backend, pack_id, expected_hash)

	response = requests.delete(backend+"/packs/"+pack_id+"/bird/robin")
	assert response.status_code == 200
	expected_hash = hashlib.sha256(
		'\x00bird\x01duck\x01eagle\x00mammal\x01cat\x01dog\x01tiger'.encode('utf-8')
	).hexdigest()
	verify_pack_hash(backend, pack_id, expected_hash)

	response = requests.delete(backend+"/packs/"+pack_id+"/mammal")
	assert response.status_code == 200
	expected_hash = hashlib.sha256(
		'\x00bird\x01duck\x01eagle'.encode('utf-8')
	).hexdigest()
	verify_pack_hash(backend, pack_id, expected_hash)

	response = requests.delete(backend+"/packs/"+pack_id+"/bird")
	assert response.status_code == 200
	expected_hash = hashlib.sha256(b'').hexdigest()
	verify_pack_hash(backend, pack_id, expected_hash)

	populate_test_pack_resources(backend, media, pack_id)
	expected_hash = hashlib.sha256(
		'\x00bird\x01duck\x01eagle\x01robin\x00mammal\x01cat\x01dog\x01tiger'.encode('utf-8')
	).hexdigest()
	verify_pack_hash(backend, pack_id, expected_hash)

	response = requests.delete(backend+"/packs/"+pack_id)
	assert response.status_code == 200


def test_pack_role_and_string_counts(backend, media):
	pack_id = create_test_pack(backend, "Test Pack Counts")
	verify_pack_counts(backend, pack_id, role_count=0, string_count=0)

	populate_test_pack_resources(backend, media, pack_id)
	verify_pack_counts(backend, pack_id, role_count=2, string_count=6)

	response = requests.delete(backend+"/packs/"+pack_id+"/mammal/cat")
	assert response.status_code == 200
	verify_pack_counts(backend, pack_id, role_count=2, string_count=5)

	response = requests.delete(backend+"/packs/"+pack_id+"/mammal/dog")
	assert response.status_code == 200
	verify_pack_counts(backend, pack_id, role_count=2, string_count=4)

	# double delete shouldn't change anything
	response = requests.delete(backend+"/packs/"+pack_id+"/mammal/dog")
	assert response.status_code == 200
	verify_pack_counts(backend, pack_id, role_count=2, string_count=4)

	# mammal role is implicitly removed because all of its string were deleted
	response = requests.delete(backend+"/packs/"+pack_id+"/mammal/tiger")
	assert response.status_code == 200
	verify_pack_counts(backend, pack_id, role_count=1, string_count=3)

	response = requests.delete(backend+"/packs/"+pack_id+"/bird")
	assert response.status_code == 200
	verify_pack_counts(backend, pack_id, role_count=0, string_count=0)

	response = requests.delete(backend+"/packs/"+pack_id)
	assert response.status_code == 200

# HELPERS

def create_test_pack(backend, title):
	pack = {"title":title}
	response = requests.post(backend + "/packs/", json=pack)
	assert response.status_code == 200
	response_data = response.json()
	pack_hash = response_data["hash"]
	# should sha256 of an empty input
	assert pack_hash == hashlib.sha256(b'').hexdigest()
	pack_id = response_data["id"]
	assert isinstance(pack_id, str)
	return pack_id

def populate_test_pack_resources(backend, media, pack_id):
	pack_api = backend + "/packs/" + pack_id
	# upload resources
	with open("./resources/bird-duck.aac", "rb") as file:
		duck_audio_id = upload_resource(pack_api+"/bird/duck", file, "audio/aac")
	with open("./resources/bird-eagle.png", "rb") as file:
		eagle_image_id = upload_resource(pack_api+"/bird/eagle", file, "image/png")
	with open("./resources/bird-eagle.mp3", "rb") as file:
		eagle_audio_id = upload_resource(pack_api+"/bird/eagle", file, "audio/mpeg")
	with open("./resources/bird-robin.jpg", "rb") as file:
		robin_image_id = upload_resource(pack_api+"/bird/robin", file, "image/jpeg")
	with open("./resources/mammal-cat.jpg", "rb") as file:
		cat_image_id = upload_resource(pack_api+"/mammal/cat", file, "image/jpeg")
	with open("./resources/mammal-cat.flac", "rb") as file:
		cat_audio_id = upload_resource(pack_api+"/mammal/cat", file, "audio/flac")
	with open("./resources/mammal-dog.webp", "rb") as file:
		dog_image_id = upload_resource(pack_api+"/mammal/dog", file, "image/webp")
	with open("./resources/mammal-tiger.svg", "rb") as file:
		tiger_image_id = upload_resource(pack_api+"/mammal/tiger", file, "image/svg+xml")
	with open("./resources/mammal-tiger.wav", "rb") as file:
		tiger_audio_id = upload_resource(pack_api+"/mammal/tiger", file, "audio/wav")
	# verify resource integrity
	with open("./resources/bird-duck.aac", "rb") as file:
		verify_resource(media+"/"+duck_audio_id, file, "audio/aac")
	with open("./resources/bird-eagle.png", "rb") as file:
		verify_resource(media+"/"+eagle_image_id, file, "image/png")
	with open("./resources/bird-eagle.mp3", "rb") as file:
		verify_resource(media+"/"+eagle_audio_id, file, "audio/mpeg")
	with open("./resources/bird-robin.jpg", "rb") as file:
		verify_resource(media+"/"+robin_image_id, file, "image/jpeg")
	with open("./resources/mammal-cat.jpg", "rb") as file:
		verify_resource(media+"/"+cat_image_id, file, "image/jpeg")
	with open("./resources/mammal-cat.flac", "rb") as file:
		verify_resource(media+"/"+cat_audio_id, file, "audio/flac")
	with open("./resources/mammal-dog.webp", "rb") as file:
		verify_resource(media+"/"+dog_image_id, file, "image/webp")
	with open("./resources/mammal-tiger.svg", "rb") as file:
		verify_resource(media+"/"+tiger_image_id, file, "image/svg+xml")
	with open("./resources/mammal-tiger.wav", "rb") as file:
		verify_resource(media+"/"+tiger_audio_id, file, "audio/wav")
	# verify resource ids in get pack api
	response = requests.get(backend+"/packs/"+pack_id)
	assert response.status_code == 200
	assert response.json()['roles'] == [
		{
			"id":"bird",
			"strings":[
				{"id":"duck","audio":duck_audio_id},
				{"id":"eagle","audio":eagle_audio_id,"image":eagle_image_id},
				{"id":"robin","image":robin_image_id}
			]
		},
		{
			"id":"mammal",
			"strings":[
				{"id":"cat","audio":cat_audio_id,"image":cat_image_id},
				{"id":"dog","image":dog_image_id},
				{"id":"tiger","audio":tiger_audio_id,"image":tiger_image_id}
			]
		}
	]

def verify_pack_hash(backend, pack_id, expected_hash):
	response = requests.get(backend+"/packs/"+pack_id)
	assert response.status_code == 200
	pack_summary = response.json()
	assert pack_summary['hash'] == expected_hash
	response = requests.get(backend+"/packs/")
	assert response.status_code == 200
	pack_list = response.json()
	assert isinstance(pack_list, list)
	filtered_data = list(filter(lambda p: p['id'] == pack_id, pack_list))
	assert len(filtered_data) == 1
	assert filtered_data[0]['hash'] == expected_hash

def verify_pack_counts(backend, pack_id, role_count, string_count):
	response = requests.get(backend+"/packs/")
	assert response.status_code == 200
	pack_list = response.json()
	assert isinstance(pack_list, list)
	filtered_data = list(filter(lambda p: p['id'] == pack_id, pack_list))
	assert len(filtered_data) == 1
	assert filtered_data[0]['roleCount'] == role_count
	assert filtered_data[0]['stringCount'] == string_count

def verify_pack_title(backend, pack_id, title):
	response = requests.get(backend+"/packs/"+pack_id)
	assert response.status_code == 200
	pack_summary = response.json()
	assert pack_summary['title'] == title
	response = requests.get(backend+"/packs/")
	assert response.status_code == 200
	pack_list = response.json()
	assert isinstance(pack_list, list)
	filtered_data = list(filter(lambda p: p['id'] == pack_id, pack_list))
	assert len(filtered_data) == 1
	assert filtered_data[0]['title'] == title

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
	resource_id = response.json()
	assert isinstance(resource_id, str)
	return resource_id

def verify_resource(url, file, content_type):
	file.seek(0)
	response = requests.get(url, stream=True)
	assert response.status_code == 200
	assert response.headers["Content-Type"] == content_type
	assert hash_response(response) == hash_file(file)

def verify_resource_deleted(url):
	response = requests.get(url)
	assert response.status_code == 404
