CREATE TABLE "NodeFavorites" (
  "user_id" integer NOT NULL DEFAULT 0,
  "device_id" integer NOT NULL,
  PRIMARY KEY ("user_id", "device_id")
);
