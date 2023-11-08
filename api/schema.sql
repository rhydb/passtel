CREATE TYPE fieldtype AS ENUM
    ('username', 'url', 'email', 'password', 'otp');

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

CREATE TABLE IF NOT EXISTS public.vault_items
(
    item_id bigint NOT NULL DEFAULT nextval('vault_items_item_id_seq'::regclass),
    vault_id bigint NOT NULL DEFAULT nextval('vault_items_vault_id_seq'::regclass),
    name character varying(20) COLLATE pg_catalog."default" NOT NULL,
    icon text COLLATE pg_catalog."default",
    CONSTRAINT vault_items_pkey PRIMARY KEY (item_id),
    CONSTRAINT vault_items_vault_id_fkey FOREIGN KEY (vault_id)
        REFERENCES public.vaults (vault_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
        NOT VALID
);

CREATE TABLE IF NOT EXISTS fields
(
    field_id bigint NOT NULL DEFAULT nextval('fields_field_url_seq'::regclass),
    item_id bigint NOT NULL DEFAULT nextval('fields_item_id_seq'::regclass),
    type fieldtype NOT NULL DEFAULT 'username'::fieldtype,
    " value" character varying COLLATE pg_catalog."default",
    CONSTRAINT fields_pkey PRIMARY KEY (field_id),
    CONSTRAINT fields_item_id_fkey FOREIGN KEY (item_id)
        REFERENCES public.vault_items (item_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
        NOT VALID
);

