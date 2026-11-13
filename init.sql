CREATE TABLE IF NOT EXISTS brands (
    BRD_iID INT UNSIGNED NOT NULL PRIMARY KEY,
    BRD_cName VARCHAR(64) NOT NULL
);

CREATE TABLE IF NOT EXISTS customers (
    CUS_iID INT UNSIGNED NOT NULL PRIMARY KEY,
    CUS_BRD_iID INT UNSIGNED NOT NULL,
    CUS_cUserName VARCHAR(64) NULL,
    CUS_cEmail VARCHAR(190) NULL,
    CUS_cFirstName VARCHAR(64) NULL,
    CUS_cLastName VARCHAR(64) NULL,
    CUS_cPhone VARCHAR(32) NULL,
    CUS_cCountry CHAR(2) NULL,
    CUS_CUR_iID INT UNSIGNED NOT NULL,
    CUS_dCreated DATETIME NOT NULL,
    CUS_bActive TINYINT(1) NOT NULL DEFAULT 1,
    UNIQUE KEY (CUS_BRD_iID, CUS_cEmail),
    KEY (CUS_cEmail),
    KEY (CUS_BRD_iID)
);

CREATE TABLE IF NOT EXISTS currencies (
    CUR_iID INT UNSIGNED NOT NULL PRIMARY KEY,
    CUR_cAlpha3 CHAR(3) NOT NULL,
    CUR_iMinorUnits TINYINT UNSIGNED NULL
);

CREATE TABLE IF NOT EXISTS transactions (
    TRX_iID BIGINT UNSIGNED NOT NULL PRIMARY KEY,
    TRX_CUS_iID INT UNSIGNED NOT NULL,
    TRX_TYP_iID INT UNSIGNED NOT NULL,
    TRX_STA_iID INT UNSIGNED NOT NULL,
    TRX_iAmount BIGINT NULL,
    TRX_PMT_iID INT UNSIGNED NULL,
    TRX_dCreated DATETIME NOT NULL,
    TRX_dBooked DATETIME NULL,
    TRX_dChanged DATETIME NOT NULL,
    KEY (TRX_CUS_iID, TRX_dCreated)
);

CREATE TABLE IF NOT EXISTS transaction_types (
    TYP_iID INT UNSIGNED NOT NULL PRIMARY KEY,
    TYP_cName VARCHAR(64) NOT NULL
);

CREATE TABLE IF NOT EXISTS transaction_statuses (
    STA_iID INT UNSIGNED NOT NULL PRIMARY KEY,
    STA_cName VARCHAR(64) NOT NULL
);

CREATE TABLE IF NOT EXISTS payment_methods (
    PMT_iID INT UNSIGNED NOT NULL PRIMARY KEY,
    PMT_cName VARCHAR(64) NOT NULL
);

CREATE TABLE IF NOT EXISTS sync_queue (
    SYQ_iID BIGINT UNSIGNED NOT NULL PRIMARY KEY,
    SYQ_CUS_iID INT UNSIGNED NOT NULL,
    SYQ_dQueued DATETIME NOT NULL,
    SYQ_dProcessed DATETIME NULL
);

CREATE TABLE IF NOT EXISTS partner_contacts (
    PCO_CUS_iID INT UNSIGNED NOT NULL PRIMARY KEY,
    PCO_cContactId VARCHAR(64) NOT NULL,
    PCO_dSynced DATETIME NOT NULL
);

-- Seed Initial Data
INSERT IGNORE INTO brands VALUES (1, 'Brand Alpha'), (2, 'Brand Beta');
INSERT IGNORE INTO currencies VALUES (1, 'EUR', 2), (2, 'GBP', 2), (3, 'JPY', 0), (4, 'BHD', 3), (5, 'USD', NULL);
INSERT IGNORE INTO transaction_types VALUES (1001, 'PAYIN'), (2001, 'PAYOUT'), (3001, 'ADJUSTMENT');
INSERT IGNORE INTO transaction_statuses VALUES 
(10, 'reserved'), (20, 'pending_provider'), (30, 'awaiting_approval'),
(40, 'approved'), (50, 'booked'), (60, 'settled'), (61, 'settled late'),
(70, 'declined_provider'), (71, 'declined risk'), (72, 'declined limit'),
(80, 'cancelled_by_customer'), (81, 'cancelled_by_operator'), (82, 'cancelled_timeout'),
(90, 'reversed'), (91, 'chargeback');
INSERT IGNORE INTO payment_methods VALUES (1, 'PAYIN CARD'), (2, 'BANK TRANSFER');
