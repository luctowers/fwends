CREATE TABLE admins (
	email varchar(255) PRIMARY KEY
);

CREATE TABLE packs (
	id bigint PRIMARY KEY,
	title varchar(255) NOT NULL
);

CREATE TYPE packResourceClass AS ENUM ('image', 'audio');

CREATE TABLE packResources (
	packId bigint,
	roleId varchar(63) NOT NULL,
	stringId varchar(63) NOT NULL,
	class packResourceClass NOT NULL,
	ready boolean NOT NULL,
	FOREIGN KEY (packId) REFERENCES packs(id)
		ON DELETE NO ACTION
  	ON UPDATE NO ACTION,
	PRIMARY KEY (packId, roleId, stringId, class)
);
