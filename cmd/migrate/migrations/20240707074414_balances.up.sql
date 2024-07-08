CREATE TABLE IF NOT EXISTS balances (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `userID` INT UNSIGNED NOT NULL,
    `tokenID` INT UNSIGNED NOT NULL,
    `amount` DECIMAL(65, 0) NOT NULL,
    FOREIGN KEY (userID) REFERENCES users(id),
    FOREIGN KEY (tokenID) REFERENCES tokens(id),
    UNIQUE KEY user_token (userID, tokenID)
);