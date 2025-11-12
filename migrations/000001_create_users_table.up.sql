CREATE TABLE "Users" (
  "user_id" serial PRIMARY KEY,
  "username" VARCHAR(50) UNIQUE NOT NULL,
  "password" VARCHAR(100) NOT NULL
);
