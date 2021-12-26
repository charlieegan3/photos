ALTER TABLE medias
ADD COLUMN exposure_time_numerator int NOT NULL DEFAULT 0,
ADD COLUMN exposure_time_denominator int NOT NULL DEFAULT 0;
