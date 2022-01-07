CREATE TABLE admins (
	email VARCHAR(255) PRIMARY KEY
);

CREATE TABLE packs (
	id BIGINT PRIMARY KEY,
	title VARCHAR(255) NOT NULL
);

CREATE TABLE packImages (
	packId BIGINT,
	roleId VARCHAR(63) NOT NULL,
	stringId VARCHAR(63) NOT NULL,
	extension VARCHAR(3) NOT NULL,
	FOREIGN KEY (packId) REFERENCES packs(id),
	PRIMARY KEY(packId, roleId, stringId)
);

CREATE TABLE packSounds (
	packId BIGINT,
	roleId VARCHAR(63) NOT NULL,
	stringId VARCHAR(63) NOT NULL,
	extension VARCHAR(3) NOT NULL,
	FOREIGN KEY (packId) REFERENCES packs(id),
	PRIMARY KEY(packId, roleId, stringId)
);
