CREATE TABLE admins (
	email varchar(255) PRIMARY KEY
);

CREATE TABLE packs (
	pack_id bigint PRIMARY KEY,
	title varchar(255) NOT NULL,
	hash bytea NOT NULL DEFAULT '\xe3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'
);
CREATE INDEX packs_hash_idx ON packs(hash);

CREATE TABLE pieces (
	hash bytea NOT NULL,
	seed bytea NOT NULL,
	title varchar(255) NOT NULL,
	PRIMARY KEY (hash, seed)
);

CREATE TABLE resources (
	resource_id bigint PRIMARY KEY
);

CREATE TABLE pruned_resources (
	resource_id bigint PRIMARY KEY
);

CREATE TYPE resourceclass AS ENUM ('image', 'audio');
CREATE TABLE pack_resources (
	pack_id bigint NOT NULL,
	role_id varchar(63) NOT NULL,
	string_id varchar(63) NOT NULL,
	resource_class resourceclass NOT NULL,
	resource_id bigint NOT NULL,
	FOREIGN KEY (pack_id) REFERENCES packs(pack_id)
		ON DELETE NO ACTION
		ON UPDATE NO ACTION,
	FOREIGN KEY (resource_id) REFERENCES resources(resource_id)
		ON DELETE NO ACTION
		ON UPDATE NO ACTION,
	PRIMARY KEY (pack_id, role_id, string_id, resource_class)
);
CREATE INDEX pack_resources_resource_id_idx ON pack_resources(resource_id);
