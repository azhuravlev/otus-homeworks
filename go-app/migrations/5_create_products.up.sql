CREATE TABLE `products` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `price` int(11) NOT NULL,
  `available` int(11) NOT NULL DEFAULT 0,
  `created_at` datetime DEFAULT NOW(),
  `updated_at` datetime DEFAULT NOW() ON UPDATE NOW(),
  PRIMARY KEY (`id`)
) DEFAULT CHARSET=utf8;
