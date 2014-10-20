-- --------------------------------------------------------

--
-- Table structure for table `clients`
--

CREATE TABLE IF NOT EXISTS `clients` (
  `nick` varchar(255) NOT NULL PRIMARY KEY,
  `channels` varchar(255) NOT NULL,
  `lastseen` int(11) NOT NULL,
) ENGINE=MyISAM DEFAULT CHARSET=utf8;

-- --------------------------------------------------------