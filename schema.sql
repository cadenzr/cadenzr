PRAGMA foreign_keys = ON;

CREATE TABLE `images` (
  `id`	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
  `path`	TEXT NOT NULL UNIQUE,
  `link`	TEXT NOT NULL UNIQUE,
  `mime` TEXT NOT NULL,
  `hash` TEXT NOT NULL UNIQUE
);
CREATE TABLE `artists` (
  `id`	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
  `name`	TEXT NOT NULL UNIQUE
);
CREATE TABLE IF NOT EXISTS "albums" (
  `id`	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
  `name`	TEXT NOT NULL UNIQUE,
  `cover_id`	INTEGER,
  `year`	INTEGER,

  FOREIGN KEY(`cover_id`) REFERENCES `images`(`id`)

);
CREATE TABLE IF NOT EXISTS "songs" (
  `id`	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
  `name`	TEXT NOT NULL,
  `artist_id`	INTEGER,
  `album_id`	INTEGER,
  `year`	INTEGER,
  `genre`	TEXT,
  `mime`	TEXT NOT NULL,
  `path`	TEXT NOT NULL,
  `cover_id`	INTEGER,
  `hash` TEXT NOT NULL,

  FOREIGN KEY(`artist_id`) REFERENCES `artists`(`id`),
  FOREIGN KEY(`album_id`) REFERENCES `albums`(`id`),
  FOREIGN KEY(`cover_id`) REFERENCES `images`(`id`)
);
