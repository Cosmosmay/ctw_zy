CREATE TABLE airbnb_info (
   id BIGINT AUTO_INCREMENT PRIMARY KEY,
   airbnb_url VARCHAR(255) NOT NULL UNIQUE, -- Reduced length and added unique constraint for URLs.
   hotel_name VARCHAR(150) NOT NULL,        -- Increased length slightly for flexibility.
   star TINYINT UNSIGNED NOT NULL,          -- `TINYINT` saves space for small numbers.
   price DECIMAL(10, 2) NOT NULL,           -- Storing price as a numeric value for calculations.
   price_before_taxes DECIMAL(10, 2) NOT NULL,
   guests SMALLINT UNSIGNED NOT NULL,       -- `SMALLINT` can handle up to 65,535 guests.
   check_in_date DATE NOT NULL,             -- Using `DATE` for check-in and check-out dates.
   check_out_date DATE NOT NULL,
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,-- Automatically set timestamp on row creation.
   PRIMARY KEY (id)
)ENGINE=InnoDB
DEFAULT CHARSET=utf8mb4
COLLATE=utf8mb4_general_ci;