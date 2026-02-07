CREATE TABLE email_verifications (
  user_id TEXT NOT NULL REFERENCES users(id),
  email VARCHAR(255) NOT NULL,
  verified_at TIMESTAMP,
  PRIMARY KEY (user_id, email)
);

CREATE TABLE email_verification_codes (
  user_id TEXT NOT NULL REFERENCES users(id),
  email VARCHAR(255) NOT NULL,
  digit_code VARCHAR(8) NOT NULL,
  valid_until TIMESTAMP NOT NULL,
  PRIMARY KEY (user_id, email, digit_code)
);
