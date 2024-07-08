CREATE TABLE IF NOT EXISTS orders (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `userID` INT UNSIGNED NOT NULL,
    `tokenID` INT UNSIGNED NOT NULL,
    `orderType` ENUM('buy', 'sell') NOT NULL,
    `amount` DECIMAL(65, 0) NOT NULL,
    `price` DECIMAL(65, 0) NOT NULL,
    `status` ENUM('open', 'filled', 'cancelled') NOT NULL DEFAULT 'open',
    `createdAt` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (userID) REFERENCES users(id),
    FOREIGN KEY (tokenID) REFERENCES tokens(id)
);