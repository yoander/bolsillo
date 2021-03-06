-- Adminer 4.3.1 MySQL dump

SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP DATABASE IF EXISTS `bolsillo`;
CREATE DATABASE `bolsillo` /*!40100 DEFAULT CHARACTER SET utf8 */;
USE `bolsillo`;

DROP TABLE IF EXISTS `invoices`;
CREATE TABLE `invoices` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(45) NOT NULL,
  `date` date NOT NULL,
  `status` enum('PAID','PNDG','CANC') NOT NULL DEFAULT 'PAID' COMMENT 'PAID = PAID, CANC = CANCELED, PNDG = PENDING',
  `note` varchar(150) NOT NULL,
  `file_path` varchar(250) NOT NULL DEFAULT '',
  `deleted` tinyint(4) NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `code_UNIQUE` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Invoices info';


DROP TABLE IF EXISTS `persons`;
CREATE TABLE `persons` (
  `id` smallint(5) unsigned NOT NULL AUTO_INCREMENT,
  `fullname` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Who belongs money transactions and invoices';


DROP TABLE IF EXISTS `tags`;
CREATE TABLE `tags` (
  `id` smallint(5) unsigned NOT NULL AUTO_INCREMENT,
  `tag` varchar(45) NOT NULL,
  `groupby` varchar(45) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `tag_UNIQUE` (`tag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Transaction clasification';


DROP TABLE IF EXISTS `transactions`;
CREATE TABLE `transactions` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `description` varchar(100) NOT NULL,
  `type` enum('EXP','INC') NOT NULL DEFAULT 'EXP' COMMENT 'EXP = Expense, INC = Income',
  `status` enum('PAID','PNDG','CANC') NOT NULL DEFAULT 'PAID',
  `price` decimal(10,5) NOT NULL DEFAULT '0.00000',
  `total_price` decimal(7,2) NOT NULL DEFAULT '0.00',
  `is_expensive` tinyint(4) NOT NULL DEFAULT '0',
  `invoice_id` int(10) unsigned DEFAULT NULL,
  `date` date NOT NULL,
  `note` varchar(150) NOT NULL DEFAULT '',
  `unit_id` tinyint(4) unsigned DEFAULT NULL,
  `quantity` decimal(7,2) NOT NULL DEFAULT '0.00',
  `person_id` smallint(5) unsigned NOT NULL,
  `deleted` tinyint(4) NOT NULL DEFAULT '0',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `fk_transaction_invoice_idx` (`invoice_id`),
  KEY `fk_transaction_person1_idx` (`person_id`),
  KEY `fk_transaction_unit1_idx` (`unit_id`),
  CONSTRAINT `fk_transaction_invoice` FOREIGN KEY (`invoice_id`) REFERENCES `invoices` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_transaction_person1` FOREIGN KEY (`person_id`) REFERENCES `persons` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_transaction_unit1` FOREIGN KEY (`unit_id`) REFERENCES `units` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Store every money transaction';


DELIMITER ;;

CREATE TRIGGER `transaction_BEFORE_INSERT` BEFORE INSERT ON `transactions` FOR EACH ROW
BEGIN
	SET @quantity = IF(NEW.quantity > 0, NEW.quantity, 1);
	/* 
     * Total price has priority over unit price 
     * NULL IS NOT > 0
     */
    IF NEW.total_price > 0 THEN
        SET NEW.price = NEW.total_price / @quantity;
    ELSE
		SET NEW.total_price = NEW.price * @quantity;
    END IF;
END;;

CREATE TRIGGER `transaction_BEFORE_UPDATE` BEFORE UPDATE ON `transactions` FOR EACH ROW
BEGIN
	SET @quantity = IF(NEW.quantity > 0, NEW.quantity, 1);
	/* 
     * Total price has priority over unit price 
     * NULL IS NOT > 0
     */
    IF NEW.total_price > 0 THEN
        SET NEW.price = NEW.total_price / @quantity;
    ELSE
		SET NEW.total_price = NEW.price * @quantity;
    END IF;
END;;

DELIMITER ;

DROP TABLE IF EXISTS `transaction_tags`;
CREATE TABLE `transaction_tags` (
  `transaction_id` int(11) unsigned NOT NULL,
  `tag_id` smallint(5) unsigned NOT NULL,
  PRIMARY KEY (`transaction_id`,`tag_id`),
  KEY `fk_transaction_has_tag_tag1_idx` (`tag_id`),
  KEY `fk_transaction_has_tag_transaction1_idx` (`transaction_id`),
  CONSTRAINT `fk_transaction_has_tag_tag1` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_transaction_has_tag_transaction1` FOREIGN KEY (`transaction_id`) REFERENCES `transactions` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


DROP TABLE IF EXISTS `units`;
CREATE TABLE `units` (
  `id` tinyint(4) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(15) NOT NULL,
  `symbol` varchar(5) NOT NULL,
  `unit_id` tinyint(4) unsigned DEFAULT NULL,
  `base_relation` decimal(10,5) unsigned DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name_UNIQUE` (`name`),
  UNIQUE KEY `symbol_UNIQUE` (`symbol`),
  KEY `fk_unit_unit1_idx` (`unit_id`),
  CONSTRAINT `fk_unit_unit1` FOREIGN KEY (`unit_id`) REFERENCES `units` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT=' International System of Units ';


-- 2018-01-05 21:11:13
