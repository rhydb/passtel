CREATE TABLE IF NOT EXISTS users (
       user_id BIGSERIAL PRIMARY KEY,
       username varchar(25) UNIQUE NOT NULL,
       password char(64) NOT NULL
);

CREATE TABLE IF NOT EXISTS tokens (
	token_id bigserial NOT NULL,
	plaintext uuid NOT NULL,
	created_at timestamp NOT NULL DEFAULT now(),
	last_used timestamp NULL,
	expires_at timestamp NULL,
	user_id bigserial NOT NULL,
	CONSTRAINT tokens_pk PRIMARY KEY (token_id),
	CONSTRAINT tokens_un UNIQUE (plaintext),
	CONSTRAINT tokens_fk FOREIGN KEY (user_id) REFERENCES public.users(user_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS vaults
(
    vault_id BIGSERIAL PRIMARY KEY,
    user_id BIGSERIAL NOT NULL,
    name varchar(50) UNIQUE NOT NULL,
    CONSTRAINT vaults_fk FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON DELETE CASCADE
        DEFERRABLE
        NOT VALID
);
