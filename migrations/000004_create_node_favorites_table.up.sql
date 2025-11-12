CREATE TABLE "NodeFavorites" (
  "user_id" integer NOT NULL,
  "device_id" integer NOT NULL,
  PRIMARY KEY ("user_id", "device_id")
);
