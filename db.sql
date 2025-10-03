CREATE TYPE "room_type" AS ENUM (
  'group',
  'direct'
);

CREATE TABLE "messages" (
  "id" SERIAL PRIMARY KEY,
  "room_id" int NOT NULL,
  "sender_id" int NOT NULL,
  "content" text NOT NULL,
  "reply_to_id" int,
  "is_pinned" bool DEFAULT false,
  "is_edited" bool DEFAULT false,
  "created_at" timestamp DEFAULT (now()),
  "updated_at" timestamp,
  "deleted_at" timestamp
);

CREATE TABLE "message_attachments" (
  "id" SERIAL PRIMARY KEY,
  "message_id" int NOT NULL,
  "file_path" text NOT NULL,
  "file_type" text NOT NULL,
  "file_size" int,
  "file_mime" text NOT NULL,
  "created_at" timestamp DEFAULT (now())
);

CREATE TABLE "rooms" (
  "id" SERIAL PRIMARY KEY,
  "name" text,
  "community_id" int NOT NULL,
  "type" room_type NOT NULL,
  "created_at" timestamp DEFAULT (now())
);

CREATE TABLE "room_participants" (
  "id" SERIAL PRIMARY KEY,
  "room_id" int NOT NULL,
  "user_id" int NOT NULL,
  "role" varchar(50) NOT NULL DEFAULT 'member',
  "joined_at" timestamp DEFAULT (now())
);

CREATE TABLE "message_reads" (
  "id" SERIAL PRIMARY KEY,
  "message_id" int NOT NULL,
  "user_id" int NOT NULL,
  "read_at" timestamp
);

CREATE TABLE "message_reactions" (
  "id" SERIAL PRIMARY KEY,
  "message_id" int NOT NULL,
  "user_id" int NOT NULL,
  "reaction_type" varchar(50) NOT NULL,
  "created_at" timestamp DEFAULT (now())
);

CREATE TABLE "communities" (
  "id" int PRIMARY KEY
);

CREATE TABLE "users" (
  "id" int PRIMARY KEY
);

CREATE TABLE "direct_message_rooms" (
  "room_id" int PRIMARY KEY,
  "community_id" int NOT NULL,
  "user_a_id" int NOT NULL,
  "user_b_id" int NOT NULL,
  CHECK (user_a_id < user_b_id),
  UNIQUE (community_id, user_a_id, user_b_id)
);

CREATE UNIQUE INDEX ON "room_participants" ("room_id", "user_id");

CREATE UNIQUE INDEX ON "message_reads" ("message_id", "user_id");

CREATE INDEX idx_messages_room_created_at ON messages (room_id, created_at DESC);

CREATE INDEX idx_messages_room_pinned ON messages (room_id, is_pinned);

CREATE INDEX idx_messages_reply_to ON messages (reply_to_id);

CREATE INDEX idx_rooms_community ON rooms (community_id);


CREATE INDEX idx_room_participants_room ON room_participants (room_id);
CREATE INDEX idx_room_participants_user ON room_participants (user_id);

CREATE INDEX idx_message_reads_user ON message_reads (user_id);


CREATE INDEX idx_message_reactions_message ON message_reactions (message_id);
CREATE INDEX idx_message_reactions_message_user ON message_reactions (message_id, user_id);
CREATE UNIQUE INDEX uniq_message_reactions_user_type ON message_reactions (message_id, user_id, reaction_type);

ALTER TABLE "messages" ADD FOREIGN KEY ("room_id") REFERENCES "rooms" ("id");

ALTER TABLE "messages" ADD FOREIGN KEY ("sender_id") REFERENCES "users" ("id");

ALTER TABLE "message_attachments" ADD FOREIGN KEY ("message_id") REFERENCES "messages" ("id");

ALTER TABLE "messages" ADD FOREIGN KEY ("reply_to_id") REFERENCES "messages" ("id");

ALTER TABLE "rooms" ADD FOREIGN KEY ("community_id") REFERENCES "communities" ("id");

ALTER TABLE "direct_message_rooms" ADD FOREIGN KEY ("room_id") REFERENCES "rooms" ("id");

ALTER TABLE "direct_message_rooms" ADD FOREIGN KEY ("community_id") REFERENCES "communities" ("id");

ALTER TABLE "direct_message_rooms" ADD FOREIGN KEY ("user_a_id") REFERENCES "users" ("id");

ALTER TABLE "direct_message_rooms" ADD FOREIGN KEY ("user_b_id") REFERENCES "users" ("id");

ALTER TABLE "room_participants" ADD FOREIGN KEY ("room_id") REFERENCES "rooms" ("id");

ALTER TABLE "room_participants" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "message_reads" ADD FOREIGN KEY ("message_id") REFERENCES "messages" ("id");

ALTER TABLE "message_reads" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

ALTER TABLE "message_reactions" ADD FOREIGN KEY ("message_id") REFERENCES "messages" ("id");

ALTER TABLE "message_reactions" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

