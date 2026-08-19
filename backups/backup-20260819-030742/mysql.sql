-- MySQL dump 10.13  Distrib 8.0.46, for Linux (x86_64)
--
-- Host: localhost    Database: dxcloud
-- ------------------------------------------------------
-- Server version	8.0.46

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `application_versions`
--

DROP TABLE IF EXISTS `application_versions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `application_versions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `application_id` bigint unsigned NOT NULL,
  `version` varchar(64) NOT NULL,
  `image_ref` varchar(255) NOT NULL,
  `commit_sha` varchar(64) NOT NULL DEFAULT '',
  `status` varchar(16) NOT NULL DEFAULT 'active' COMMENT 'active/superseded/failed',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_app_versions` (`application_id`,`version`),
  KEY `idx_app_versions_app` (`application_id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='应用版本';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `application_versions`
--

LOCK TABLES `application_versions` WRITE;
/*!40000 ALTER TABLE `application_versions` DISABLE KEYS */;
INSERT INTO `application_versions` VALUES (1,1,'2.8','registry:2.8','','active','2026-08-19 01:13:51.296'),(3,1,'2.8.7','registry:2.8','','active','2026-08-19 01:17:11.561'),(4,3,'15000/default/pipetest:v5.13','host.docker.internal:15000/default/pipetest:v5','','active','2026-08-19 01:49:47.869'),(5,3,'15000/default/pipetest:v5.14','host.docker.internal:15000/default/pipetest:v5','','active','2026-08-19 01:50:11.899'),(6,3,'15000/default/pipetest:v5.15','host.docker.internal:15000/default/pipetest:v5','','active','2026-08-19 02:08:23.281');
/*!40000 ALTER TABLE `application_versions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `applications`
--

DROP TABLE IF EXISTS `applications`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `applications` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned DEFAULT NULL,
  `project_id` bigint unsigned DEFAULT NULL,
  `owner_id` bigint unsigned NOT NULL DEFAULT '0',
  `name` varchar(64) NOT NULL,
  `type` varchar(32) NOT NULL DEFAULT 'custom' COMMENT 'node/go/java/python/nginx/mysql/redis/postgres/custom',
  `image` varchar(255) NOT NULL DEFAULT '' COMMENT '当前/默认镜像（部署时可覆盖）',
  `git_url` varchar(512) NOT NULL DEFAULT '',
  `git_branch` varchar(64) NOT NULL DEFAULT 'main',
  `port` int NOT NULL DEFAULT '80',
  `health_check_path` varchar(255) NOT NULL DEFAULT '',
  `env` text COMMENT '默认环境变量 JSON',
  `domain` varchar(255) NOT NULL DEFAULT '',
  `active_deployment_id` bigint unsigned DEFAULT NULL,
  `status` tinyint NOT NULL DEFAULT '1',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_apps_project_name` (`project_id`,`name`),
  KEY `idx_apps_project` (`project_id`),
  KEY `idx_apps_owner` (`owner_id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='应用（PaaS）';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `applications`
--

LOCK TABLES `applications` WRITE;
/*!40000 ALTER TABLE `applications` DISABLE KEYS */;
INSERT INTO `applications` VALUES (1,NULL,1,1,'registry-app','custom','registry:2.8','','',5000,'/v2/','[]','app1.localhost',7,1,'2026-08-19 01:06:59.032','2026-08-19 01:17:13.915','2026-08-19 01:17:13.915'),(2,NULL,1,1,'bad-app','custom','registry:2.8','','',5000,'/nonexistent-path','[]','bad.localhost',NULL,1,'2026-08-19 01:09:06.510','2026-08-19 01:17:13.939','2026-08-19 01:17:13.939'),(3,NULL,NULL,1,'pipe-app','custom','registry:2.8','','',80,'/','[]','pipe.localhost',15,1,'2026-08-19 01:34:17.992','2026-08-19 02:08:23.287',NULL),(4,3,4,1,'appa-0819023310','container','busybox:latest','','',80,'','null','',NULL,1,'2026-08-19 02:33:15.086','2026-08-19 02:33:15.086',NULL);
/*!40000 ALTER TABLE `applications` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `audit_logs`
--

DROP TABLE IF EXISTS `audit_logs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `audit_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned DEFAULT NULL,
  `user_id` bigint unsigned DEFAULT NULL,
  `action` varchar(64) NOT NULL,
  `resource_type` varchar(32) NOT NULL DEFAULT '',
  `resource_id` varchar(64) NOT NULL DEFAULT '',
  `ip` varchar(64) NOT NULL DEFAULT '',
  `request_id` varchar(64) NOT NULL DEFAULT '',
  `detail` text,
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '1=成功 2=拒绝',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_audit_org` (`org_id`,`created_at`),
  KEY `idx_audit_user` (`user_id`,`created_at`),
  KEY `idx_audit_resource` (`resource_type`,`resource_id`)
) ENGINE=InnoDB AUTO_INCREMENT=279 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='审计日志';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `audit_logs`
--

LOCK TABLES `audit_logs` WRITE;
/*!40000 ALTER TABLE `audit_logs` DISABLE KEYS */;
INSERT INTO `audit_logs` VALUES (1,NULL,2,'auth.register','user','2','172.21.0.1','req-afcbf7ef','',1,'2026-08-18 23:46:51.612'),(2,NULL,1,'auth.login','user','1','172.21.0.1','req-dd551aee','',1,'2026-08-18 23:46:51.693'),(3,NULL,NULL,'auth.login','user','admin','172.21.0.1','req-e36bf16d','{\"username\":\"admin\"}',2,'2026-08-18 23:46:51.771'),(4,NULL,1,'auth.login','user','1','172.21.0.1','req-e915616c','',1,'2026-08-18 23:47:40.262'),(5,NULL,2,'auth.login','user','2','172.21.0.1','req-42ffe470','',1,'2026-08-18 23:49:13.900'),(6,NULL,2,'authz.deny','api','/api/v1/users','172.21.0.1','req-44032f15','{\"perm\":\"user:list\"}',2,'2026-08-18 23:49:13.953'),(7,NULL,2,'auth.login','user','2','172.21.0.1','req-b44fd12f','',1,'2026-08-18 23:52:09.777'),(8,NULL,2,'auth.login','user','2','172.21.0.1','req-91887da0','',1,'2026-08-18 23:55:36.838'),(9,NULL,2,'auth.login','user','2','172.21.0.1','req-93c7dec5','',1,'2026-08-18 23:55:36.949'),(10,NULL,2,'authz.deny','api','/api/v1/users','172.21.0.1','req-aacc459f','{\"perm\":\"user:list\"}',2,'2026-08-18 23:55:36.985'),(11,NULL,2,'auth.login','user','2','172.21.0.1','req-6ac56049','',1,'2026-08-18 23:55:52.776'),(12,NULL,2,'auth.logout','user','2','172.21.0.1','req-c0194581','',1,'2026-08-18 23:55:52.800'),(13,NULL,1,'auth.login','user','1','172.21.0.1','req-3e9f0d08','',1,'2026-08-18 23:55:52.889'),(14,NULL,1,'auth.login','user','1','172.21.0.1','req-9ed59229','',1,'2026-08-19 00:13:14.098'),(15,NULL,1,'ecs.create','ecs','i-0d9adf92cf6c6861','172.21.0.1','req-46fa334f','{\"image\":\"alpine:3.20\",\"name\":\"test-web\"}',1,'2026-08-19 00:13:14.150'),(16,NULL,1,'ecs.stop','ecs','i-0d9adf92cf6c6861','172.21.0.1','req-d02c48f8','',1,'2026-08-19 00:13:27.573'),(17,NULL,1,'ecs.start','ecs','i-0d9adf92cf6c6861','172.21.0.1','req-a39910be','',1,'2026-08-19 00:13:27.717'),(18,NULL,1,'ecs.restart','ecs','i-0d9adf92cf6c6861','172.21.0.1','req-25fd9b62','',1,'2026-08-19 00:13:38.008'),(19,NULL,1,'ecs.force-stop','ecs','i-0d9adf92cf6c6861','172.21.0.1','req-62c15008','',1,'2026-08-19 00:13:38.177'),(20,NULL,1,'auth.login','user','1','172.21.0.1','req-12052828','',1,'2026-08-19 00:13:54.185'),(21,NULL,1,'ecs.create','ecs','i-091d68b04e97d6ad','172.21.0.1','req-54982b51','{\"image\":\"alpine:3.20\",\"name\":\"port-a\"}',1,'2026-08-19 00:13:54.287'),(22,NULL,2,'auth.login','user','2','172.21.0.1','req-86ba0bc8','',1,'2026-08-19 00:13:54.687'),(23,NULL,2,'auth.login','user','2','172.21.0.1','req-0b63aeb6','',1,'2026-08-19 00:14:22.691'),(24,NULL,2,'ecs.create','ecs','i-0bcb08fc56dbbb76','172.21.0.1','req-47237c45','{\"image\":\"alpine:3.20\",\"name\":\"q1\"}',1,'2026-08-19 00:14:22.712'),(25,NULL,2,'ecs.create','ecs','i-f2f2188374bda892','172.21.0.1','req-bb222ed1','{\"image\":\"alpine:3.20\",\"name\":\"q2\"}',1,'2026-08-19 00:14:22.909'),(26,NULL,2,'ecs.create','ecs','i-07accab39b9b5872','172.21.0.1','req-5d8cd52e','{\"image\":\"alpine:3.20\",\"name\":\"q3\"}',1,'2026-08-19 00:14:23.123'),(27,NULL,2,'ecs.create','ecs','i-b4633bb999715b99','172.21.0.1','req-862d1694','{\"image\":\"alpine:3.20\",\"name\":\"q4\"}',1,'2026-08-19 00:14:23.346'),(28,NULL,2,'ecs.create','ecs','i-d3d78d7475201067','172.21.0.1','req-7d2047f8','{\"image\":\"alpine:3.20\",\"name\":\"q5\"}',1,'2026-08-19 00:14:23.543'),(29,NULL,2,'ecs.delete','ecs','i-0bcb08fc56dbbb76','172.21.0.1','req-152cec21','',1,'2026-08-19 00:15:08.995'),(30,NULL,1,'auth.login','user','1','172.21.0.1','req-d9a73501','',1,'2026-08-19 00:15:19.931'),(31,NULL,1,'ecs.delete','ecs','i-0d9adf92cf6c6861','172.21.0.1','req-752000a6','',1,'2026-08-19 00:15:19.966'),(32,NULL,1,'ecs.delete','ecs','i-091d68b04e97d6ad','172.21.0.1','req-74c3d182','',1,'2026-08-19 00:15:20.235'),(33,NULL,1,'ecs.delete','ecs','i-f2f2188374bda892','172.21.0.1','req-19fca538','',1,'2026-08-19 00:15:20.408'),(34,NULL,1,'ecs.delete','ecs','i-07accab39b9b5872','172.21.0.1','req-bdc3196c','',1,'2026-08-19 00:15:20.630'),(35,NULL,1,'ecs.delete','ecs','i-b4633bb999715b99','172.21.0.1','req-543e994f','',1,'2026-08-19 00:15:20.861'),(36,NULL,1,'ecs.delete','ecs','i-d3d78d7475201067','172.21.0.1','req-7fe4ec73','',1,'2026-08-19 00:15:21.102'),(37,NULL,1,'auth.login','user','1','172.21.0.1','req-f706d977','',1,'2026-08-19 00:24:56.041'),(38,NULL,1,'ecs.create','ecs','i-f5b41e50f40f57cb','172.21.0.1','req-9bc2cb82','{\"image\":\"alpine:3.20\",\"name\":\"term-test\"}',1,'2026-08-19 00:24:56.084'),(39,NULL,1,'console.token','ecs','8','172.21.0.1','req-4f46b670','',1,'2026-08-19 00:24:56.291'),(40,NULL,1,'console.open','ecs','i-f5b41e50f40f57cb','172.21.0.1','','',1,'2026-08-19 00:24:56.320'),(41,NULL,1,'console.close','ecs','i-f5b41e50f40f57cb','172.21.0.1','','{\"duration_sec\":10.032526059}',1,'2026-08-19 00:25:06.356'),(42,NULL,1,'auth.login','user','1','172.21.0.1','req-c47cd642','',1,'2026-08-19 00:27:36.560'),(43,NULL,1,'console.token','ecs','8','172.21.0.1','req-a4286794','',1,'2026-08-19 00:27:36.605'),(44,NULL,1,'console.open','ecs','i-f5b41e50f40f57cb','172.21.0.1','','',1,'2026-08-19 00:27:36.634'),(45,NULL,1,'console.close','ecs','i-f5b41e50f40f57cb','172.21.0.1','','{\"duration_sec\":2.420867566}',1,'2026-08-19 00:27:39.060'),(46,NULL,1,'auth.login','user','1','172.21.0.1','req-8367897a','',1,'2026-08-19 00:29:09.843'),(47,NULL,1,'console.token','ecs','8','172.21.0.1','req-9b636027','',1,'2026-08-19 00:29:09.876'),(48,NULL,1,'console.open','ecs','i-f5b41e50f40f57cb','172.21.0.1','','',1,'2026-08-19 00:29:09.893'),(49,NULL,1,'console.close','ecs','i-f5b41e50f40f57cb','172.21.0.1','','{\"duration_sec\":119.907138068}',1,'2026-08-19 00:31:09.801'),(50,NULL,1,'auth.login','user','1','172.21.0.1','req-4728eedf','',1,'2026-08-19 00:32:13.597'),(51,NULL,1,'console.token','ecs','8','172.21.0.1','req-6e9788d6','',1,'2026-08-19 00:32:13.625'),(52,NULL,1,'console.open','ecs','i-f5b41e50f40f57cb','172.21.0.1','','',1,'2026-08-19 00:32:13.641'),(53,NULL,1,'console.close','ecs','i-f5b41e50f40f57cb','172.21.0.1','','{\"duration_sec\":0.157566697}',1,'2026-08-19 00:32:13.802'),(54,NULL,1,'auth.login','user','1','172.21.0.1','req-92edf2e5','',1,'2026-08-19 00:32:35.880'),(55,NULL,1,'console.token','ecs','8','172.21.0.1','req-88879351','',1,'2026-08-19 00:32:35.892'),(56,NULL,1,'console.open','ecs','i-f5b41e50f40f57cb','172.21.0.1','','',1,'2026-08-19 00:32:35.908'),(57,NULL,1,'console.close','ecs','i-f5b41e50f40f57cb','172.21.0.1','','{\"duration_sec\":89.943235731}',1,'2026-08-19 00:34:05.851'),(58,NULL,1,'auth.login','user','1','172.21.0.1','req-795dadbc','',1,'2026-08-19 00:35:29.071'),(59,NULL,1,'console.token','ecs','8','172.21.0.1','req-a277b940','',1,'2026-08-19 00:35:29.085'),(60,NULL,1,'console.open','ecs','i-f5b41e50f40f57cb','172.21.0.1','','',1,'2026-08-19 00:35:29.099'),(61,NULL,1,'console.close','ecs','i-f5b41e50f40f57cb','172.21.0.1','','{\"duration_sec\":5.009054965}',1,'2026-08-19 00:35:34.113'),(62,NULL,1,'auth.login','user','1','172.21.0.1','req-ee2c48b5','',1,'2026-08-19 00:35:49.027'),(63,NULL,1,'console.token','ecs','8','172.21.0.1','req-5b3fb07d','',1,'2026-08-19 00:35:49.040'),(64,NULL,1,'console.open','ecs','i-f5b41e50f40f57cb','172.21.0.1','','',1,'2026-08-19 00:35:49.053'),(65,NULL,1,'console.close','ecs','i-f5b41e50f40f57cb','172.21.0.1','','{\"duration_sec\":5.005748632}',1,'2026-08-19 00:35:54.063'),(66,NULL,1,'auth.login','user','1','172.21.0.1','req-b16889f3','',1,'2026-08-19 00:36:06.955'),(67,NULL,2,'auth.login','user','2','172.21.0.1','req-28697ac9','',1,'2026-08-19 00:36:07.056'),(68,NULL,1,'ecs.delete','ecs','i-f5b41e50f40f57cb','172.21.0.1','req-9efc5dad','',1,'2026-08-19 00:36:07.601'),(69,NULL,1,'auth.login','user','1','172.21.0.1','req-11e0bcdc','',1,'2026-08-19 00:48:12.622'),(70,NULL,1,'image.pull','image','hello-world:latest','172.21.0.1','req-b4afdb1c','',1,'2026-08-19 00:48:12.668'),(71,NULL,1,'image.tag','image','hello-world:latest -> registry:5000/default/hello:v1','172.21.0.1','req-83d55419','',1,'2026-08-19 00:48:17.865'),(72,NULL,1,'auth.login','user','1','172.21.0.1','req-dd669eeb','',1,'2026-08-19 00:48:32.239'),(73,NULL,1,'ecs.create','ecs','i-edadff776dd52cd8','172.21.0.1','req-c438ea8c','{\"image\":\"alpine:3.20\",\"name\":\"net-web\"}',1,'2026-08-19 00:48:32.710'),(74,NULL,1,'ecs.create','ecs','i-9d21eed2d323b228','172.21.0.1','req-e6ec64d9','{\"image\":\"alpine:3.20\",\"name\":\"net-db\"}',1,'2026-08-19 00:48:35.100'),(75,NULL,1,'ecs.delete','ecs','i-edadff776dd52cd8','172.21.0.1','req-44b8e88e','',1,'2026-08-19 00:48:35.626'),(76,NULL,1,'ecs.delete','ecs','i-9d21eed2d323b228','172.21.0.1','req-528c1fb8','',1,'2026-08-19 00:48:35.822'),(77,NULL,1,'auth.login','user','1','172.21.0.1','req-91c5fb62','',1,'2026-08-19 00:49:51.369'),(78,NULL,1,'network.create','network','net-a','172.21.0.1','req-5f2c2718','{\"subnet\":\"10.30.0.0/24\"}',1,'2026-08-19 00:49:51.458'),(79,NULL,1,'ecs.create','ecs','i-0c924f733cb9c034','172.21.0.1','req-36a52074','{\"image\":\"alpine:3.20\",\"name\":\"net-web\"}',1,'2026-08-19 00:49:51.497'),(80,NULL,1,'ecs.create','ecs','i-09f5134845cb9ae9','172.21.0.1','req-b97bc934','{\"image\":\"alpine:3.20\",\"name\":\"net-db\"}',1,'2026-08-19 00:49:51.880'),(81,NULL,1,'network.connect','network','net-a','172.21.0.1','req-a0b0b544','{\"instance\":\"i-09f5134845cb9ae9\",\"ip\":\"\"}',1,'2026-08-19 00:49:52.164'),(82,NULL,1,'network.disconnect','network','net-a','172.21.0.1','req-69dc30fa','{\"instance\":\"i-09f5134845cb9ae9\"}',1,'2026-08-19 00:49:52.334'),(83,NULL,1,'ecs.delete','ecs','i-0c924f733cb9c034','172.21.0.1','req-23e17b2b','',1,'2026-08-19 00:49:52.504'),(84,NULL,1,'ecs.delete','ecs','i-09f5134845cb9ae9','172.21.0.1','req-bad5531e','',1,'2026-08-19 00:49:52.687'),(85,NULL,1,'network.delete','network','net-a','172.21.0.1','req-d0a55505','',1,'2026-08-19 00:49:52.974'),(86,NULL,1,'auth.login','user','1','172.21.0.1','req-eb50b3ec','',1,'2026-08-19 00:51:08.942'),(87,NULL,1,'volume.create','volume','app-data','172.21.0.1','req-ac3b10df','{\"capacity_gb\":10}',1,'2026-08-19 00:51:09.009'),(88,NULL,1,'ecs.create','ecs','i-9beedd088196d217','172.21.0.1','req-e363b667','{\"image\":\"alpine:3.20\",\"name\":\"vol-web\"}',1,'2026-08-19 00:51:09.048'),(89,NULL,1,'volume.create','volume','app-data2','172.21.0.1','req-19bda2c9','{\"capacity_gb\":5}',1,'2026-08-19 00:51:09.514'),(90,NULL,1,'volume.attach','ecs','i-9beedd088196d217','172.21.0.1','req-8f9ea17d','{\"target\":\"/logs\",\"volume\":\"app-data2\"}',1,'2026-08-19 00:51:09.911'),(91,NULL,1,'volume.detach','ecs','i-9beedd088196d217','172.21.0.1','req-d6feb197','{\"volume\":\"app-data2\"}',1,'2026-08-19 00:51:10.494'),(92,NULL,1,'ecs.delete','ecs','i-9beedd088196d217','172.21.0.1','req-76570e98','',1,'2026-08-19 00:51:10.718'),(93,NULL,1,'volume.delete','volume','app-data','172.21.0.1','req-73dedf47','',1,'2026-08-19 00:51:10.744'),(94,NULL,1,'volume.delete','volume','app-data2','172.21.0.1','req-3d51eca6','',1,'2026-08-19 00:51:10.774'),(95,NULL,1,'auth.login','user','1','172.21.0.1','req-1a5e0f3a','',1,'2026-08-19 00:51:21.702'),(96,NULL,1,'auth.login','user','1','172.21.0.1','req-fb5cd77e','',1,'2026-08-19 00:53:06.303'),(97,NULL,1,'registry.delete_tag','registry','default/app:v1','172.21.0.1','req-af96088c','',1,'2026-08-19 00:53:18.974'),(98,NULL,1,'auth.login','user','1','172.21.0.1','req-e369a804','',1,'2026-08-19 00:55:00.539'),(99,NULL,1,'registry.delete_tag','registry','default/app:v1','172.21.0.1','req-ffaab3c4','',1,'2026-08-19 00:55:01.680'),(100,NULL,1,'auth.login','user','1','172.21.0.1','req-42df9f98','',1,'2026-08-19 00:57:02.855'),(101,NULL,1,'registry.pull','registry','default/app:v1','172.21.0.1','req-d8d71f60','',1,'2026-08-19 00:57:03.975'),(102,NULL,1,'registry.delete_tag','registry','default/app:v1','172.21.0.1','req-964ae01d','',1,'2026-08-19 00:57:04.024'),(103,NULL,1,'auth.login','user','1','172.21.0.1','req-eeda5654','',1,'2026-08-19 01:06:58.927'),(104,NULL,1,'project.create','project','shop','172.21.0.1','req-a934630c','',1,'2026-08-19 01:06:59.011'),(105,NULL,1,'app.create','app','registry-app','172.21.0.1','req-0ae41268','',1,'2026-08-19 01:06:59.035'),(106,NULL,1,'deploy.start','app','registry-app','172.21.0.1','req-69a2622f','{\"deployment\":1,\"image\":\"registry:2.8\"}',1,'2026-08-19 01:06:59.082'),(107,NULL,1,'deploy.failed','app','registry-app','172.21.0.1','req-69a2622f','{\"deployment\":1}',2,'2026-08-19 01:07:59.610'),(108,NULL,1,'deploy.start','app','registry-app','172.21.0.1','req-19f6fb00','{\"deployment\":2,\"image\":\"registry:2.8\"}',1,'2026-08-19 01:08:02.866'),(109,NULL,1,'deploy.failed','app','registry-app','172.21.0.1','req-19f6fb00','{\"deployment\":2}',2,'2026-08-19 01:09:03.356'),(110,NULL,1,'app.create','app','bad-app','172.21.0.1','req-07e142ff','',1,'2026-08-19 01:09:06.513'),(111,NULL,1,'deploy.start','app','bad-app','172.21.0.1','req-a21d65e7','{\"deployment\":3,\"image\":\"registry:2.8\"}',1,'2026-08-19 01:09:06.559'),(112,NULL,1,'deploy.failed','app','bad-app','172.21.0.1','req-a21d65e7','{\"deployment\":3}',2,'2026-08-19 01:10:07.093'),(113,NULL,1,'auth.login','user','1','172.21.0.1','req-d665beba','',1,'2026-08-19 01:11:33.108'),(114,NULL,1,'deploy.start','app','registry-app','172.21.0.1','req-e99524aa','{\"deployment\":4,\"image\":\"registry:2.8\"}',1,'2026-08-19 01:11:33.182'),(115,NULL,1,'deploy.failed','app','registry-app','172.21.0.1','req-e99524aa','{\"deployment\":4}',2,'2026-08-19 01:12:33.680'),(116,NULL,1,'auth.login','user','1','172.21.0.1','req-718d48e1','',1,'2026-08-19 01:13:50.917'),(117,NULL,1,'deploy.start','app','registry-app','172.21.0.1','req-209019c1','{\"deployment\":5,\"image\":\"registry:2.8\"}',1,'2026-08-19 01:13:51.026'),(118,NULL,1,'deploy.success','app','registry-app','172.21.0.1','req-209019c1','{\"deployment\":5,\"image\":\"registry:2.8\",\"version\":\"2.8\"}',1,'2026-08-19 01:13:51.310'),(119,NULL,1,'auth.login','user','1','172.21.0.1','req-4175dac3','',1,'2026-08-19 01:15:51.116'),(120,NULL,1,'deploy.start','app','registry-app','172.21.0.1','req-b1f12f94','{\"deployment\":6,\"image\":\"registry:2.8\"}',1,'2026-08-19 01:15:51.204'),(121,NULL,1,'deploy.success','app','registry-app','172.21.0.1','req-b1f12f94','{\"deployment\":6,\"image\":\"registry:2.8\",\"version\":\"2.8\"}',1,'2026-08-19 01:15:52.051'),(122,NULL,1,'auth.login','user','1','172.21.0.1','req-6535e110','',1,'2026-08-19 01:17:10.339'),(123,NULL,1,'deploy.start','app','registry-app','172.21.0.1','req-d21c4c3c','{\"deployment\":7,\"image\":\"registry:2.8\"}',1,'2026-08-19 01:17:10.558'),(124,NULL,1,'deploy.success','app','registry-app','172.21.0.1','req-d21c4c3c','{\"deployment\":7,\"image\":\"registry:2.8\",\"version\":\"2.8\"}',1,'2026-08-19 01:17:11.576'),(125,NULL,1,'app.delete','app','registry-app','172.21.0.1','req-9cfce318','',1,'2026-08-19 01:17:13.920'),(126,NULL,1,'app.delete','app','bad-app','172.21.0.1','req-8a89df7c','',1,'2026-08-19 01:17:13.945'),(127,NULL,1,'auth.login','user','1','172.21.0.1','req-ca4dcb24','',1,'2026-08-19 01:24:32.800'),(128,NULL,1,'pipeline.create','pipeline','hello-pipe','172.21.0.1','req-ac77afe9','',1,'2026-08-19 01:24:32.835'),(129,NULL,1,'pipeline.run','pipeline','hello-pipe','172.21.0.1','req-606988b1','{\"run\":1}',1,'2026-08-19 01:24:32.871'),(130,NULL,1,'auth.login','user','1','172.21.0.1','req-8d854758','',1,'2026-08-19 01:24:52.992'),(131,NULL,1,'pipeline.create','pipeline','fail-pipe','172.21.0.1','req-a5f1b0e1','',1,'2026-08-19 01:24:53.023'),(132,NULL,1,'pipeline.run','pipeline','fail-pipe','172.21.0.1','req-1dba08d8','{\"run\":2}',1,'2026-08-19 01:24:53.065'),(133,NULL,1,'pipeline.create','pipeline','cancel-pipe','172.21.0.1','req-055adf6a','',1,'2026-08-19 01:24:58.104'),(134,NULL,1,'pipeline.run','pipeline','cancel-pipe','172.21.0.1','req-37720f77','{\"run\":3}',1,'2026-08-19 01:24:58.124'),(135,NULL,1,'pipeline.cancel','pipeline-run','3','172.21.0.1','req-55c9b33e','',1,'2026-08-19 01:25:06.259'),(136,NULL,1,'auth.login','user','1','172.21.0.1','req-9e6197a6','',1,'2026-08-19 01:26:30.695'),(137,NULL,1,'pipeline.run','pipeline','cancel-pipe','172.21.0.1','req-3d5d74d7','{\"run\":4}',1,'2026-08-19 01:26:30.744'),(138,NULL,1,'pipeline.cancel','pipeline-run','4','172.21.0.1','req-3a439b22','',1,'2026-08-19 01:26:38.877'),(139,NULL,1,'pipeline.run','pipeline','fail-pipe','172.21.0.1','req-25a7ac6c','{\"run\":5}',1,'2026-08-19 01:26:43.917'),(140,NULL,1,'auth.login','user','1','172.21.0.1','req-a1903b68','',1,'2026-08-19 01:28:00.396'),(141,NULL,1,'pipeline.run','pipeline','cancel-pipe','172.21.0.1','req-18dbd94e','{\"run\":6}',1,'2026-08-19 01:28:00.441'),(142,NULL,1,'pipeline.cancel','pipeline-run','6','172.21.0.1','req-448e3ff3','',1,'2026-08-19 01:28:08.596'),(143,NULL,1,'auth.login','user','1','172.21.0.1','req-5e9ffb81','',1,'2026-08-19 01:34:17.955'),(144,NULL,1,'app.create','app','pipe-app','172.21.0.1','req-14d945cf','',1,'2026-08-19 01:34:17.994'),(145,NULL,1,'pipeline.create','pipeline','ci-cd','172.21.0.1','req-0c1812ed','',1,'2026-08-19 01:34:18.009'),(146,NULL,1,'webhook.create','webhook','github/aea857b1dad67536','172.21.0.1','req-96a9a3d3','',1,'2026-08-19 01:34:18.019'),(147,NULL,1,'auth.login','user','1','172.21.0.1','req-5329bd52','',1,'2026-08-19 01:34:28.621'),(148,NULL,1,'pipeline.run','pipeline','ci-cd','172.21.0.1','req-3f15681d','{\"run\":7}',1,'2026-08-19 01:34:28.679'),(149,NULL,1,'auth.login','user','1','172.21.0.1','req-8cb55666','',1,'2026-08-19 01:34:44.746'),(150,NULL,1,'auth.login','user','1','172.21.0.1','req-806595f1','',1,'2026-08-19 01:36:15.753'),(151,NULL,1,'pipeline.run','pipeline','ci-cd','172.21.0.1','req-f174296d','{\"run\":8}',1,'2026-08-19 01:36:15.808'),(152,NULL,1,'auth.login','user','1','172.21.0.1','req-53fbcd31','',1,'2026-08-19 01:37:36.051'),(153,NULL,1,'pipeline.run','pipeline','ci-cd','172.21.0.1','req-5411e2bf','{\"run\":9}',1,'2026-08-19 01:37:36.104'),(154,NULL,1,'auth.login','user','1','172.21.0.1','req-def39606','',1,'2026-08-19 01:37:47.613'),(155,NULL,1,'auth.login','user','1','172.21.0.1','req-c2215432','',1,'2026-08-19 01:39:03.108'),(156,NULL,1,'pipeline.run','pipeline','ci-cd','172.21.0.1','req-3572b36f','{\"run\":10}',1,'2026-08-19 01:39:03.159'),(157,NULL,1,'deploy.start','app','pipe-app','','','{\"deployment\":8,\"image\":\"host.docker.internal:15000/default/pipetest:v1\"}',1,'2026-08-19 01:39:04.750'),(158,NULL,1,'auth.login','user','1','172.21.0.1','req-39b6fb80','',1,'2026-08-19 01:39:14.683'),(159,NULL,1,'auth.login','user','1','172.21.0.1','req-fd47322f','',1,'2026-08-19 01:40:19.034'),(160,NULL,1,'pipeline.run','pipeline','ci-cd','172.21.0.1','req-0296a2d8','{\"run\":11}',1,'2026-08-19 01:40:19.087'),(161,NULL,1,'deploy.start','app','pipe-app','','','{\"deployment\":9,\"image\":\"host.docker.internal:15000/default/pipetest:v1\"}',1,'2026-08-19 01:40:19.987'),(162,NULL,1,'auth.login','user','1','172.21.0.1','req-e8314543','',1,'2026-08-19 01:41:16.110'),(163,NULL,1,'pipeline.update','pipeline','ci-cd','172.21.0.1','req-0ab0cd03','',1,'2026-08-19 01:41:16.142'),(164,NULL,1,'pipeline.run','pipeline','ci-cd','172.21.0.1','req-c22d7e09','{\"run\":12}',1,'2026-08-19 01:41:16.167'),(165,NULL,1,'deploy.start','app','pipe-app','','','{\"deployment\":10,\"image\":\"host.docker.internal:15000/default/pipetest:v2\"}',1,'2026-08-19 01:41:17.192'),(166,NULL,1,'deploy.failed','app','pipe-app','','','{\"deployment\":10}',2,'2026-08-19 01:42:17.674'),(167,NULL,1,'auth.login','user','1','172.21.0.1','req-fe2770c9','',1,'2026-08-19 01:44:11.847'),(168,NULL,1,'pipeline.update','pipeline','ci-cd','172.21.0.1','req-20c17f8b','',1,'2026-08-19 01:44:11.880'),(169,NULL,1,'pipeline.run','pipeline','ci-cd','172.21.0.1','req-495fb0e4','{\"run\":13}',1,'2026-08-19 01:44:11.911'),(170,NULL,1,'deploy.start','app','pipe-app','','','{\"deployment\":11,\"image\":\"host.docker.internal:15000/default/pipetest:v3\"}',1,'2026-08-19 01:44:13.012'),(171,NULL,1,'deploy.failed','app','pipe-app','','','{\"deployment\":11}',2,'2026-08-19 01:45:13.497'),(172,NULL,1,'auth.login','user','1','172.21.0.1','req-c792bd80','',1,'2026-08-19 01:46:06.633'),(173,NULL,1,'auth.login','user','1','172.21.0.1','req-e57a7d80','',1,'2026-08-19 01:46:42.405'),(174,NULL,1,'pipeline.create','pipeline','debug-cat','172.21.0.1','req-3b78e576','',1,'2026-08-19 01:46:42.421'),(175,NULL,1,'pipeline.run','pipeline','debug-cat','172.21.0.1','req-0e5092ff','{\"run\":14}',1,'2026-08-19 01:46:42.461'),(176,NULL,1,'auth.login','user','1','172.21.0.1','req-531e9d29','',1,'2026-08-19 01:47:03.150'),(177,NULL,1,'pipeline.create','pipeline','debug-cat2','172.21.0.1','req-e41b088a','',1,'2026-08-19 01:47:03.165'),(178,NULL,1,'pipeline.run','pipeline','debug-cat2','172.21.0.1','req-f0b7d243','{\"run\":15}',1,'2026-08-19 01:47:03.204'),(179,NULL,1,'auth.login','user','1','172.21.0.1','req-3e0f0f88','',1,'2026-08-19 01:48:03.634'),(180,NULL,1,'pipeline.update','pipeline','ci-cd','172.21.0.1','req-fb6e6f8d','',1,'2026-08-19 01:48:03.667'),(181,NULL,1,'pipeline.run','pipeline','ci-cd','172.21.0.1','req-c815f280','{\"run\":16}',1,'2026-08-19 01:48:03.692'),(182,NULL,1,'deploy.start','app','pipe-app','','','{\"deployment\":12,\"image\":\"host.docker.internal:15000/default/pipetest:v4\"}',1,'2026-08-19 01:48:04.818'),(183,NULL,1,'deploy.failed','app','pipe-app','','','{\"deployment\":12}',2,'2026-08-19 01:49:05.364'),(184,NULL,1,'auth.login','user','1','172.21.0.1','req-c1b3942d','',1,'2026-08-19 01:49:46.341'),(185,NULL,1,'pipeline.update','pipeline','ci-cd','172.21.0.1','req-42e17fd3','',1,'2026-08-19 01:49:46.375'),(186,NULL,1,'pipeline.run','pipeline','ci-cd','172.21.0.1','req-b06c18db','{\"run\":17}',1,'2026-08-19 01:49:46.406'),(187,NULL,1,'deploy.start','app','pipe-app','','','{\"deployment\":13,\"image\":\"host.docker.internal:15000/default/pipetest:v5\"}',1,'2026-08-19 01:49:47.658'),(188,NULL,1,'deploy.success','app','pipe-app','','','{\"deployment\":13,\"image\":\"host.docker.internal:15000/default/pipetest:v5\",\"version\":\"15000/default/pipetest:v5\"}',1,'2026-08-19 01:49:47.882'),(189,NULL,1,'auth.login','user','1','172.21.0.1','req-ad3fdef8','',1,'2026-08-19 01:50:10.059'),(190,NULL,1,'webhook.create','webhook','github/d42811c5227ca84f','172.21.0.1','req-8b9be6bf','',1,'2026-08-19 01:50:10.115'),(191,NULL,1,'pipeline.run','pipeline','ci-cd','172.21.0.1','req-20bef86f','{\"run\":18}',1,'2026-08-19 01:50:10.158'),(192,NULL,1,'webhook.delete','webhook','2','172.21.0.1','req-e31dfc76','',1,'2026-08-19 01:50:10.171'),(193,NULL,1,'deploy.start','app','pipe-app','','','{\"deployment\":14,\"image\":\"host.docker.internal:15000/default/pipetest:v5\"}',1,'2026-08-19 01:50:11.160'),(194,NULL,1,'deploy.success','app','pipe-app','','','{\"deployment\":14,\"image\":\"host.docker.internal:15000/default/pipetest:v5\",\"version\":\"15000/default/pipetest:v5\"}',1,'2026-08-19 01:50:11.911'),(195,NULL,1,'auth.login','user','1','172.21.0.1','req-a88e2534','',1,'2026-08-19 01:50:21.132'),(196,NULL,1,'pipeline.delete','pipeline','6','172.21.0.1','req-c8c0f82f','',1,'2026-08-19 01:50:21.153'),(197,NULL,1,'pipeline.delete','pipeline','5','172.21.0.1','req-1014a7cc','',1,'2026-08-19 01:50:21.162'),(198,NULL,1,'pipeline.delete','pipeline','3','172.21.0.1','req-ede4221a','',1,'2026-08-19 01:50:21.181'),(199,NULL,1,'pipeline.delete','pipeline','2','172.21.0.1','req-24ab5dd1','',1,'2026-08-19 01:50:21.190'),(200,NULL,1,'pipeline.delete','pipeline','1','172.21.0.1','req-caa35dae','',1,'2026-08-19 01:50:21.198'),(201,NULL,1,'auth.login','user','1','172.21.0.1','req-60afcdeb','',1,'2026-08-19 02:05:40.659'),(202,NULL,1,'ecs.create','ecs','i-673c58dbdabe0455','172.21.0.1','req-f16dbad8','{\"image\":\"alpine:3.20\",\"name\":\"mon-demo\"}',1,'2026-08-19 02:05:40.718'),(203,NULL,1,'auth.login','user','1','172.21.0.1','req-821dcf07','',1,'2026-08-19 02:08:11.345'),(204,NULL,1,'ecs.stop','ecs','i-673c58dbdabe0455','172.21.0.1','req-d66dcec8','',1,'2026-08-19 02:08:21.535'),(205,NULL,1,'ecs.start','ecs','i-673c58dbdabe0455','172.21.0.1','req-f8895abe','',1,'2026-08-19 02:08:21.680'),(206,NULL,1,'pipeline.run','pipeline','ci-cd','172.21.0.1','req-c2d10eeb','{\"run\":19}',1,'2026-08-19 02:08:21.750'),(207,NULL,1,'deploy.start','app','pipe-app','','','{\"deployment\":15,\"image\":\"host.docker.internal:15000/default/pipetest:v5\"}',1,'2026-08-19 02:08:22.619'),(208,NULL,1,'deploy.success','app','pipe-app','','','{\"deployment\":15,\"image\":\"host.docker.internal:15000/default/pipetest:v5\",\"version\":\"15000/default/pipetest:v5\"}',1,'2026-08-19 02:08:23.290'),(209,NULL,1,'auth.login','user','1','172.21.0.1','req-836fd7e1','',1,'2026-08-19 02:08:34.151'),(210,NULL,1,'ecs.delete','ecs','i-673c58dbdabe0455','172.21.0.1','req-8ede1d15','',1,'2026-08-19 02:08:34.359'),(211,NULL,1,'auth.login','user','1','172.21.0.1','req-71eb5ec2','',1,'2026-08-19 02:28:45.866'),(212,NULL,1,'auth.login','user','1','172.21.0.1','req-c279cf59','',1,'2026-08-19 02:29:50.639'),(213,NULL,1,'org.create','org','1','172.21.0.1','req-2628f325','{\"name\":\"??A-0819022950\"}',1,'2026-08-19 02:29:50.781'),(214,NULL,1,'org.create','org','2','172.21.0.1','req-dd47566f','{\"name\":\"??B-0819022950\"}',1,'2026-08-19 02:29:50.841'),(215,NULL,3,'auth.register','user','3','172.21.0.1','req-6fb78834','',1,'2026-08-19 02:29:50.958'),(216,NULL,3,'auth.login','user','3','172.21.0.1','req-bc36c326','',1,'2026-08-19 02:29:51.063'),(217,NULL,1,'org.member.add','org','1','172.21.0.1','req-ded46197','{\"role\":\"member\",\"user\":\"bob0819022950\"}',1,'2026-08-19 02:29:51.089'),(218,NULL,1,'project.create','project','pa-web-0819022950','172.21.0.1','req-ed4957fb','',1,'2026-08-19 02:29:51.181'),(219,NULL,1,'project.create','project','pb-web-0819022950','172.21.0.1','req-8d329fb3','',1,'2026-08-19 02:29:51.232'),(220,NULL,1,'ecs.create','ecs','i-8868a5136c00b7d3','172.21.0.1','req-a0206591','{\"image\":\"busybox:latest\",\"name\":\"tenant-a-ecs1\"}',1,'2026-08-19 02:29:51.396'),(221,NULL,1,'billing.tick','billing','-','172.21.0.1','req-18de3615','',1,'2026-08-19 02:29:54.855'),(222,NULL,1,'billing.recharge','billing','-','172.21.0.1','req-f467edae','{\"amount\":500,\"org\":1}',1,'2026-08-19 02:29:54.914'),(223,NULL,1,'billing.recharge','billing','-','172.21.0.1','req-c6318e24','{\"amount\":5000,\"org\":1}',1,'2026-08-19 02:29:55.187'),(224,NULL,1,'ecs.create','ecs','i-f82fbf823e70dee8','172.21.0.1','req-a26d3d1d','{\"image\":\"busybox:latest\",\"name\":\"tenant-a-ecs3\"}',1,'2026-08-19 02:29:55.210'),(225,NULL,1,'auth.login','user','1','172.21.0.1','req-bd07aa64','',1,'2026-08-19 02:30:36.784'),(226,NULL,1,'auth.login','user','1','172.21.0.1','req-c27679fd','',1,'2026-08-19 02:31:56.743'),(227,NULL,1,'auth.login','user','1','172.21.0.1','req-ab3c7133','',1,'2026-08-19 02:32:06.013'),(228,NULL,1,'auth.login','user','1','172.21.0.1','req-38a77948','',1,'2026-08-19 02:32:12.956'),(229,NULL,1,'auth.login','user','1','172.21.0.1','req-69c0f91c','',1,'2026-08-19 02:33:10.837'),(230,NULL,1,'org.create','org','3','172.21.0.1','req-fe10fa84','{\"name\":\"??A-0819023310\"}',1,'2026-08-19 02:33:10.896'),(231,NULL,1,'org.create','org','4','172.21.0.1','req-25ce30c6','{\"name\":\"??B-0819023310\"}',1,'2026-08-19 02:33:10.935'),(232,NULL,4,'auth.register','user','4','172.21.0.1','req-a2cfb3e7','',1,'2026-08-19 02:33:11.016'),(233,NULL,4,'auth.login','user','4','172.21.0.1','req-fe2a2dd1','',1,'2026-08-19 02:33:11.085'),(234,NULL,1,'org.member.add','org','3','172.21.0.1','req-51f37524','{\"role\":\"member\",\"user\":\"bob0819023310\"}',1,'2026-08-19 02:33:11.116'),(235,NULL,1,'project.create','project','pa-web-0819023310','172.21.0.1','req-8e7291e6','',1,'2026-08-19 02:33:11.166'),(236,NULL,1,'project.create','project','pb-web-0819023310','172.21.0.1','req-ea0d7ef5','',1,'2026-08-19 02:33:11.201'),(237,NULL,1,'ecs.create','ecs','i-79258b28f6e449cf','172.21.0.1','req-035c0967','{\"image\":\"busybox:latest\",\"name\":\"tenant-a-ecs1\"}',1,'2026-08-19 02:33:11.299'),(238,NULL,1,'billing.tick','billing','-','172.21.0.1','req-3ff12449','',1,'2026-08-19 02:33:14.587'),(239,NULL,1,'billing.recharge','billing','-','172.21.0.1','req-04f1c406','{\"amount\":500,\"org\":3}',1,'2026-08-19 02:33:14.625'),(240,NULL,1,'billing.recharge','billing','-','172.21.0.1','req-e552d09c','{\"amount\":5000,\"org\":3}',1,'2026-08-19 02:33:14.830'),(241,NULL,1,'ecs.create','ecs','i-f2769eac89fcc103','172.21.0.1','req-97f0d3cc','{\"image\":\"busybox:latest\",\"name\":\"tenant-a-ecs3\"}',1,'2026-08-19 02:33:14.847'),(242,NULL,1,'app.create','app','appa-0819023310','172.21.0.1','req-eb0871bc','',1,'2026-08-19 02:33:15.095'),(243,NULL,1,'auth.login','user','1','172.21.0.1','req-efaa401a','',1,'2026-08-19 02:36:32.189'),(244,NULL,1,'auth.login','user','1','172.21.0.1','req-21c96e83','',1,'2026-08-19 02:43:14.350'),(245,NULL,1,'security.scan','security','-','172.21.0.1','req-936ab675','',1,'2026-08-19 02:43:14.665'),(246,NULL,1,'secret.create','secret','DB_PASSWORD_P11','172.21.0.1','req-9f6270ba','',1,'2026-08-19 02:43:14.726'),(247,NULL,1,'secret.reveal','secret','1','172.21.0.1','req-5dfd8c75','',1,'2026-08-19 02:43:14.768'),(248,NULL,1,'secret.delete','secret','1','172.21.0.1','req-f7439e86','',1,'2026-08-19 02:43:14.814'),(249,NULL,5,'auth.register','user','5','172.21.0.1','req-410e1510','',1,'2026-08-19 02:43:14.896'),(250,NULL,NULL,'auth.login','user','locku0819024314','172.21.0.1','req-3608c004','{\"username\":\"locku0819024314\"}',2,'2026-08-19 02:43:14.968'),(251,NULL,NULL,'auth.login','user','locku0819024314','172.21.0.1','req-c088763d','{\"username\":\"locku0819024314\"}',2,'2026-08-19 02:43:15.040'),(252,NULL,NULL,'auth.login','user','locku0819024314','172.21.0.1','req-5834e134','{\"username\":\"locku0819024314\"}',2,'2026-08-19 02:43:15.107'),(253,NULL,NULL,'auth.login','user','locku0819024314','172.21.0.1','req-3c4adb6b','{\"username\":\"locku0819024314\"}',2,'2026-08-19 02:43:15.174'),(254,NULL,5,'auth.login','user','5','172.21.0.1','req-af87c1f2','',1,'2026-08-19 02:44:17.517'),(255,NULL,1,'auth.login','user','1','172.21.0.1','req-571a002c','',1,'2026-08-19 02:44:31.756'),(256,NULL,1,'security.scan','security','-','172.21.0.1','req-9536da34','',1,'2026-08-19 02:44:32.028'),(257,NULL,1,'secret.create','secret','DB_PASSWORD_P11','172.21.0.1','req-4d9799d8','',1,'2026-08-19 02:44:32.075'),(258,NULL,1,'secret.reveal','secret','2','172.21.0.1','req-60b418fa','',1,'2026-08-19 02:44:32.110'),(259,NULL,1,'secret.delete','secret','2','172.21.0.1','req-1c964afa','',1,'2026-08-19 02:44:32.159'),(260,NULL,6,'auth.register','user','6','172.21.0.1','req-f6146a83','',1,'2026-08-19 02:44:32.245'),(261,NULL,NULL,'auth.login','user','locku0819024432','172.21.0.1','req-388d8403','{\"username\":\"locku0819024432\"}',2,'2026-08-19 02:44:32.314'),(262,NULL,NULL,'auth.login','user','locku0819024432','172.21.0.1','req-8b1a6279','{\"username\":\"locku0819024432\"}',2,'2026-08-19 02:44:32.388'),(263,NULL,NULL,'auth.login','user','locku0819024432','172.21.0.1','req-c204627d','{\"username\":\"locku0819024432\"}',2,'2026-08-19 02:44:32.458'),(264,NULL,6,'auth.login','user','6','172.21.0.1','req-db4725fc','',1,'2026-08-19 02:45:34.806'),(265,NULL,1,'auth.login','user','1','172.21.0.1','req-acf489ed','',1,'2026-08-19 02:45:48.363'),(266,NULL,1,'security.scan','security','-','172.21.0.1','req-d60d0845','',1,'2026-08-19 02:45:48.650'),(267,NULL,1,'secret.create','secret','DB_PASSWORD_P11','172.21.0.1','req-2de20fbe','',1,'2026-08-19 02:45:48.700'),(268,NULL,1,'secret.reveal','secret','3','172.21.0.1','req-41bd9a6f','',1,'2026-08-19 02:45:48.733'),(269,NULL,1,'secret.delete','secret','3','172.21.0.1','req-a9a9e342','',1,'2026-08-19 02:45:48.776'),(270,NULL,7,'auth.register','user','7','172.21.0.1','req-04417f77','',1,'2026-08-19 02:46:53.877'),(271,NULL,NULL,'auth.login','user','locku0819024653','172.21.0.1','req-9f00d348','{\"username\":\"locku0819024653\"}',2,'2026-08-19 02:46:53.946'),(272,NULL,NULL,'auth.login','user','locku0819024653','172.21.0.1','req-9545e722','{\"username\":\"locku0819024653\"}',2,'2026-08-19 02:46:54.014'),(273,NULL,NULL,'auth.login','user','locku0819024653','172.21.0.1','req-3120cd87','{\"username\":\"locku0819024653\"}',2,'2026-08-19 02:46:54.080'),(274,NULL,NULL,'auth.login','user','locku0819024653','172.21.0.1','req-f97c0900','{\"username\":\"locku0819024653\"}',2,'2026-08-19 02:46:54.233'),(275,NULL,7,'auth.login','user','7','172.21.0.1','req-95963782','',1,'2026-08-19 02:47:56.460'),(276,NULL,1,'auth.login','user','1','172.21.0.1','req-10daa1c8','',1,'2026-08-19 03:06:49.500'),(277,NULL,1,'auth.login','user','1','172.21.0.1','req-4148d9ad','',1,'2026-08-19 03:07:37.730'),(278,NULL,1,'ecs.create','ecs','i-3c155604e70975b7','172.21.0.1','req-50ec209d','{\"image\":\"busybox:latest\",\"name\":\"prod-smoke-ecs\"}',1,'2026-08-19 03:07:37.827');
/*!40000 ALTER TABLE `audit_logs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `deployments`
--

DROP TABLE IF EXISTS `deployments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `deployments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned DEFAULT NULL,
  `project_id` bigint unsigned DEFAULT NULL,
  `application_id` bigint unsigned NOT NULL,
  `environment_id` bigint unsigned DEFAULT NULL,
  `version_id` bigint unsigned DEFAULT NULL,
  `version` varchar(64) NOT NULL DEFAULT '',
  `image_ref` varchar(255) NOT NULL,
  `strategy` varchar(16) NOT NULL DEFAULT 'blue-green',
  `status` varchar(16) NOT NULL DEFAULT 'pending' COMMENT 'pending/deploying/success/failed/rolled-back',
  `health_status` varchar(16) NOT NULL DEFAULT '' COMMENT 'healthy/unhealthy',
  `trigger_type` varchar(16) NOT NULL DEFAULT 'manual' COMMENT 'manual/webhook/pipeline',
  `pipeline_run_id` bigint unsigned DEFAULT NULL,
  `container_id` varchar(128) NOT NULL DEFAULT '',
  `container_name` varchar(128) NOT NULL DEFAULT '',
  `config_json` text COMMENT '容器配置快照（用于降级重建/回滚）',
  `note` varchar(255) NOT NULL DEFAULT '',
  `started_at` datetime(3) DEFAULT NULL,
  `finished_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_deployments_app` (`application_id`,`status`),
  KEY `idx_deployments_project` (`project_id`)
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='部署记录';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `deployments`
--

LOCK TABLES `deployments` WRITE;
/*!40000 ALTER TABLE `deployments` DISABLE KEYS */;
INSERT INTO `deployments` VALUES (1,NULL,1,1,NULL,NULL,'2.8','registry:2.8','blue-green','failed','unhealthy','manual',NULL,'20631c4cea99f2854f154edd7c424e880dd3bd9dc1cf772f092c2fbce357fee2','dx-app-1-v1','{\"env\":[],\"image\":\"registry:2.8\",\"labels\":{\"com.dxcloud.app-id\":\"1\",\"com.dxcloud.deploy-id\":\"1\",\"com.dxcloud.kind\":\"app\",\"traefik.enable\":\"true\",\"traefik.http.routers.app1-v1.entrypoints\":\"web\",\"traefik.http.routers.app1-v1.priority\":\"10\",\"traefik.http.routers.app1-v1.rule\":\"Host(`app1.localhost`)\",\"traefik.http.routers.app1-v1.service\":\"app1-v1\",\"traefik.http.services.app1-v1.loadbalancer.server.port\":\"5000\"},\"name\":\"dx-app-1-v1\",\"port\":5000,\"ports\":[],\"router_key\":\"app1-v1\"}','健康检查未通过','2026-08-19 01:06:59.077','2026-08-19 01:07:59.602','2026-08-19 01:06:59.077','2026-08-19 01:07:59.602'),(2,NULL,1,1,NULL,NULL,'2.8','registry:2.8','blue-green','failed','unhealthy','manual',NULL,'b4fc034f3b68f58c715fd4139d969076dd0db5da08ac71f69ddc4a90350717c0','dx-app-1-v2','{\"env\":[\"VERSION=v2\"],\"image\":\"registry:2.8\",\"labels\":{\"com.dxcloud.app-id\":\"1\",\"com.dxcloud.deploy-id\":\"2\",\"com.dxcloud.kind\":\"app\",\"traefik.enable\":\"true\",\"traefik.http.routers.app1-v2.entrypoints\":\"web\",\"traefik.http.routers.app1-v2.priority\":\"10\",\"traefik.http.routers.app1-v2.rule\":\"Host(`app1.localhost`)\",\"traefik.http.routers.app1-v2.service\":\"app1-v2\",\"traefik.http.services.app1-v2.loadbalancer.server.port\":\"5000\"},\"name\":\"dx-app-1-v2\",\"port\":5000,\"ports\":[],\"router_key\":\"app1-v2\"}','健康检查未通过','2026-08-19 01:08:02.862','2026-08-19 01:09:03.346','2026-08-19 01:08:02.862','2026-08-19 01:09:03.347'),(3,NULL,1,2,NULL,NULL,'2.8','registry:2.8','blue-green','failed','unhealthy','manual',NULL,'03c0a90dc546847987b19abeea5243b3394ef0d110f9266d6e26a368550aa79a','dx-app-2-v3','{\"env\":[],\"image\":\"registry:2.8\",\"labels\":{\"com.dxcloud.app-id\":\"2\",\"com.dxcloud.deploy-id\":\"3\",\"com.dxcloud.kind\":\"app\",\"traefik.enable\":\"true\",\"traefik.http.routers.app2-v3.entrypoints\":\"web\",\"traefik.http.routers.app2-v3.priority\":\"10\",\"traefik.http.routers.app2-v3.rule\":\"Host(`bad.localhost`)\",\"traefik.http.routers.app2-v3.service\":\"app2-v3\",\"traefik.http.services.app2-v3.loadbalancer.server.port\":\"5000\"},\"name\":\"dx-app-2-v3\",\"port\":5000,\"ports\":[],\"router_key\":\"app2-v3\"}','健康检查未通过','2026-08-19 01:09:06.554','2026-08-19 01:10:07.084','2026-08-19 01:09:06.555','2026-08-19 01:10:07.085'),(4,NULL,1,1,NULL,NULL,'2.8','registry:2.8','blue-green','failed','unhealthy','manual',NULL,'d3f559f4282c49e141425f3297fa400038c2c19660f842006c19b8aa52634d98','dx-app-1-v4','{\"env\":[],\"image\":\"registry:2.8\",\"labels\":{\"com.dxcloud.app-id\":\"1\",\"com.dxcloud.deploy-id\":\"4\",\"com.dxcloud.kind\":\"app\",\"traefik.enable\":\"true\",\"traefik.http.routers.app1-v4.entrypoints\":\"web\",\"traefik.http.routers.app1-v4.priority\":\"10\",\"traefik.http.routers.app1-v4.rule\":\"Host(`app1.localhost`)\",\"traefik.http.routers.app1-v4.service\":\"app1-v4\",\"traefik.http.services.app1-v4.loadbalancer.server.port\":\"5000\"},\"name\":\"dx-app-1-v4\",\"port\":5000,\"ports\":[],\"router_key\":\"app1-v4\"}','健康检查未通过','2026-08-19 01:11:33.178','2026-08-19 01:12:33.674','2026-08-19 01:11:33.178','2026-08-19 01:12:33.675'),(5,NULL,1,1,NULL,NULL,'2.8','registry:2.8','blue-green','success','healthy','manual',NULL,'7a5928f73d12308bc4c5334fc802f04390d7eec0a53101f95b02bdce5dd710a6','dx-app-1-v5','{\"env\":[],\"image\":\"registry:2.8\",\"labels\":{\"com.dxcloud.app-id\":\"1\",\"com.dxcloud.deploy-id\":\"5\",\"com.dxcloud.kind\":\"app\",\"traefik.enable\":\"true\",\"traefik.http.routers.app1-v5.entrypoints\":\"web\",\"traefik.http.routers.app1-v5.priority\":\"10\",\"traefik.http.routers.app1-v5.rule\":\"Host(`app1.localhost`)\",\"traefik.http.routers.app1-v5.service\":\"app1-v5\",\"traefik.http.services.app1-v5.loadbalancer.server.port\":\"5000\"},\"name\":\"dx-app-1-v5\",\"port\":5000,\"ports\":[],\"router_key\":\"app1-v5\"}','v1','2026-08-19 01:13:51.021','2026-08-19 01:13:51.300','2026-08-19 01:13:51.021','2026-08-19 01:15:52.037'),(6,NULL,1,1,NULL,NULL,'2.8','registry:2.8','blue-green','success','healthy','manual',NULL,'53db0f5504fe930db23426d84a102c1567728e76b6d9d4f107420bd315796137','dx-app-1-v6','{\"env\":[],\"image\":\"registry:2.8\",\"labels\":{\"com.dxcloud.app-id\":\"1\",\"com.dxcloud.deploy-id\":\"6\",\"com.dxcloud.kind\":\"app\",\"traefik.enable\":\"true\",\"traefik.http.routers.app1-v6.entrypoints\":\"web\",\"traefik.http.routers.app1-v6.priority\":\"20\",\"traefik.http.routers.app1-v6.rule\":\"Host(`app1.localhost`)\",\"traefik.http.routers.app1-v6.service\":\"app1-v6\",\"traefik.http.services.app1-v6.loadbalancer.server.port\":\"5000\"},\"name\":\"dx-app-1-v6\",\"port\":5000,\"ports\":[],\"router_key\":\"app1-v6\"}','v6-priority20','2026-08-19 01:15:51.200','2026-08-19 01:15:52.043','2026-08-19 01:15:51.200','2026-08-19 01:17:11.555'),(7,NULL,1,1,NULL,NULL,'2.8','registry:2.8','blue-green','success','healthy','manual',NULL,'9bc7677f1d4dbb11aedd242e2b79aebc72a4016553d59fe4acf9c31c096ce94c','dx-app-1-v7','{\"env\":[],\"image\":\"registry:2.8\",\"labels\":{\"com.dxcloud.app-id\":\"1\",\"com.dxcloud.deploy-id\":\"7\",\"com.dxcloud.kind\":\"app\",\"traefik.enable\":\"true\",\"traefik.http.routers.app1-v7.entrypoints\":\"web\",\"traefik.http.routers.app1-v7.priority\":\"20\",\"traefik.http.routers.app1-v7.rule\":\"Host(`app1.localhost`)\",\"traefik.http.routers.app1-v7.service\":\"app1-v7\",\"traefik.http.services.app1-v7.loadbalancer.server.port\":\"5000\"},\"name\":\"dx-app-1-v7\",\"port\":5000,\"ports\":[],\"router_key\":\"app1-v7\"}','rollback to 2.8','2026-08-19 01:17:10.540','2026-08-19 01:17:11.566','2026-08-19 01:17:10.540','2026-08-19 01:17:11.566'),(8,NULL,NULL,3,NULL,NULL,'15000/default/pipetest:v1','host.docker.internal:15000/default/pipetest:v1','blue-green','deploying','','pipeline',NULL,'','','','pipeline run #10','2026-08-19 01:39:04.745',NULL,'2026-08-19 01:39:04.745','2026-08-19 01:39:04.745'),(9,NULL,NULL,3,NULL,NULL,'15000/default/pipetest:v1','host.docker.internal:15000/default/pipetest:v1','blue-green','deploying','','pipeline',NULL,'','','','pipeline run #11','2026-08-19 01:40:19.982',NULL,'2026-08-19 01:40:19.983','2026-08-19 01:40:19.983'),(10,NULL,NULL,3,NULL,NULL,'15000/default/pipetest:v2','host.docker.internal:15000/default/pipetest:v2','blue-green','failed','unhealthy','pipeline',NULL,'734c7719452dca55f58c1f8680573b91dded1ab80bb23056a5e258e8d42ac1e0','dx-app-3-v10','{\"env\":[],\"image\":\"host.docker.internal:15000/default/pipetest:v2\",\"labels\":{\"com.dxcloud.app-id\":\"3\",\"com.dxcloud.deploy-id\":\"10\",\"com.dxcloud.kind\":\"app\",\"traefik.enable\":\"true\",\"traefik.http.routers.app3-v10.entrypoints\":\"web\",\"traefik.http.routers.app3-v10.priority\":\"20\",\"traefik.http.routers.app3-v10.rule\":\"Host(`pipe.localhost`)\",\"traefik.http.routers.app3-v10.service\":\"app3-v10\",\"traefik.http.services.app3-v10.loadbalancer.server.port\":\"80\"},\"name\":\"dx-app-3-v10\",\"port\":80,\"ports\":[],\"router_key\":\"app3-v10\"}','健康检查未通过','2026-08-19 01:41:17.189','2026-08-19 01:42:17.669','2026-08-19 01:41:17.189','2026-08-19 01:42:17.669'),(11,NULL,NULL,3,NULL,NULL,'15000/default/pipetest:v3','host.docker.internal:15000/default/pipetest:v3','blue-green','failed','unhealthy','pipeline',NULL,'a9eadccb36e7c4dcfa26e78874865945ae680e6e14cc3ebd0cf253cf7d7c4cd2','dx-app-3-v11','{\"env\":[],\"image\":\"host.docker.internal:15000/default/pipetest:v3\",\"labels\":{\"com.dxcloud.app-id\":\"3\",\"com.dxcloud.deploy-id\":\"11\",\"com.dxcloud.kind\":\"app\",\"traefik.enable\":\"true\",\"traefik.http.routers.app3-v11.entrypoints\":\"web\",\"traefik.http.routers.app3-v11.priority\":\"20\",\"traefik.http.routers.app3-v11.rule\":\"Host(`pipe.localhost`)\",\"traefik.http.routers.app3-v11.service\":\"app3-v11\",\"traefik.http.services.app3-v11.loadbalancer.server.port\":\"80\"},\"name\":\"dx-app-3-v11\",\"port\":80,\"ports\":[],\"router_key\":\"app3-v11\"}','健康检查未通过','2026-08-19 01:44:13.008','2026-08-19 01:45:13.491','2026-08-19 01:44:13.009','2026-08-19 01:45:13.492'),(12,NULL,NULL,3,NULL,NULL,'15000/default/pipetest:v4','host.docker.internal:15000/default/pipetest:v4','blue-green','failed','unhealthy','pipeline',NULL,'2af9a505d285b06de8d32472809dad8475d937843d01fceddcb3026fef0fbc82','dx-app-3-v12','{\"env\":[],\"image\":\"host.docker.internal:15000/default/pipetest:v4\",\"labels\":{\"com.dxcloud.app-id\":\"3\",\"com.dxcloud.deploy-id\":\"12\",\"com.dxcloud.kind\":\"app\",\"traefik.enable\":\"true\",\"traefik.http.routers.app3-v12.entrypoints\":\"web\",\"traefik.http.routers.app3-v12.priority\":\"20\",\"traefik.http.routers.app3-v12.rule\":\"Host(`pipe.localhost`)\",\"traefik.http.routers.app3-v12.service\":\"app3-v12\",\"traefik.http.services.app3-v12.loadbalancer.server.port\":\"80\"},\"name\":\"dx-app-3-v12\",\"port\":80,\"ports\":[],\"router_key\":\"app3-v12\"}','健康检查未通过','2026-08-19 01:48:04.813','2026-08-19 01:49:05.358','2026-08-19 01:48:04.814','2026-08-19 01:49:05.358'),(13,NULL,NULL,3,NULL,NULL,'15000/default/pipetest:v5','host.docker.internal:15000/default/pipetest:v5','blue-green','success','healthy','pipeline',NULL,'1970b5e0b756a1150dfb7e9fa3808b4f2044c83c8757f0be66ba562e1d9c149b','dx-app-3-v13','{\"env\":[],\"image\":\"host.docker.internal:15000/default/pipetest:v5\",\"labels\":{\"com.dxcloud.app-id\":\"3\",\"com.dxcloud.deploy-id\":\"13\",\"com.dxcloud.kind\":\"app\",\"traefik.enable\":\"true\",\"traefik.http.routers.app3-v13.entrypoints\":\"web\",\"traefik.http.routers.app3-v13.priority\":\"20\",\"traefik.http.routers.app3-v13.rule\":\"Host(`pipe.localhost`)\",\"traefik.http.routers.app3-v13.service\":\"app3-v13\",\"traefik.http.services.app3-v13.loadbalancer.server.port\":\"80\"},\"name\":\"dx-app-3-v13\",\"port\":80,\"ports\":[],\"router_key\":\"app3-v13\"}','pipeline run #17','2026-08-19 01:49:47.654','2026-08-19 01:49:47.874','2026-08-19 01:49:47.655','2026-08-19 01:50:11.895'),(14,NULL,NULL,3,NULL,NULL,'15000/default/pipetest:v5','host.docker.internal:15000/default/pipetest:v5','blue-green','success','healthy','pipeline',NULL,'28ace4a92e24a3c16d85636b20767b8497cc705f2ec53e1558a0630f0ab434c0','dx-app-3-v14','{\"env\":[],\"image\":\"host.docker.internal:15000/default/pipetest:v5\",\"labels\":{\"com.dxcloud.app-id\":\"3\",\"com.dxcloud.deploy-id\":\"14\",\"com.dxcloud.kind\":\"app\",\"traefik.enable\":\"true\",\"traefik.http.routers.app3-v14.entrypoints\":\"web\",\"traefik.http.routers.app3-v14.priority\":\"20\",\"traefik.http.routers.app3-v14.rule\":\"Host(`pipe.localhost`)\",\"traefik.http.routers.app3-v14.service\":\"app3-v14\",\"traefik.http.services.app3-v14.loadbalancer.server.port\":\"80\"},\"name\":\"dx-app-3-v14\",\"port\":80,\"ports\":[],\"router_key\":\"app3-v14\"}','pipeline run #18','2026-08-19 01:50:11.154','2026-08-19 01:50:11.903','2026-08-19 01:50:11.155','2026-08-19 02:08:23.276'),(15,NULL,NULL,3,NULL,NULL,'15000/default/pipetest:v5','host.docker.internal:15000/default/pipetest:v5','blue-green','success','healthy','pipeline',NULL,'c521507acc3e34a76b3a414e36918ef0dc403dbbc22e466d0e2682e2a7563eee','dx-app-3-v15','{\"env\":[],\"image\":\"host.docker.internal:15000/default/pipetest:v5\",\"labels\":{\"com.dxcloud.app-id\":\"3\",\"com.dxcloud.deploy-id\":\"15\",\"com.dxcloud.kind\":\"app\",\"traefik.enable\":\"true\",\"traefik.http.routers.app3-v15.entrypoints\":\"web\",\"traefik.http.routers.app3-v15.priority\":\"20\",\"traefik.http.routers.app3-v15.rule\":\"Host(`pipe.localhost`)\",\"traefik.http.routers.app3-v15.service\":\"app3-v15\",\"traefik.http.services.app3-v15.loadbalancer.server.port\":\"80\"},\"name\":\"dx-app-3-v15\",\"port\":80,\"ports\":[],\"router_key\":\"app3-v15\"}','pipeline run #19','2026-08-19 02:08:22.613','2026-08-19 02:08:23.283','2026-08-19 02:08:22.614','2026-08-19 02:08:23.284');
/*!40000 ALTER TABLE `deployments` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `docker_images`
--

DROP TABLE IF EXISTS `docker_images`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `docker_images` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned DEFAULT NULL,
  `project_id` bigint unsigned DEFAULT NULL,
  `repo` varchar(255) NOT NULL COMMENT 'repository 名（含 registry 前缀）',
  `tag` varchar(128) NOT NULL DEFAULT 'latest',
  `image_id` varchar(128) NOT NULL DEFAULT '' COMMENT 'digest/sha',
  `size_bytes` bigint NOT NULL DEFAULT '0',
  `docker_created_at` datetime(3) DEFAULT NULL,
  `source` varchar(16) NOT NULL DEFAULT 'pull' COMMENT 'pull/build/import',
  `status` varchar(16) NOT NULL DEFAULT 'ready' COMMENT 'pulling/ready/failed',
  `pull_error` varchar(512) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_docker_images` (`repo`,`tag`),
  KEY `idx_docker_images_org` (`org_id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Docker 镜像（镜像中心）';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `docker_images`
--

LOCK TABLES `docker_images` WRITE;
/*!40000 ALTER TABLE `docker_images` DISABLE KEYS */;
INSERT INTO `docker_images` VALUES (1,NULL,NULL,'hello-world','latest','sha256:5dd0d3e6e255913fc30f90b9f2b1d359cc2cbdb48090cc4b65f1676e203243cc',15077,'2026-03-24 05:33:59.562','pull','ready','','2026-08-19 00:48:12.664','2026-08-19 00:48:17.033',NULL),(2,NULL,NULL,'registry:5000/default/hello','v1','sha256:5dd0d3e6e255913fc30f90b9f2b1d359cc2cbdb48090cc4b65f1676e203243cc',15077,'2026-03-24 05:33:59.562','import','ready','','2026-08-19 00:48:17.860','2026-08-19 00:48:17.860',NULL),(3,NULL,NULL,'registry:5000/default/app','v1','sha256:d1a8d0a4eeb63aff09f5f34d4d80505e0ba81905f36158cc3970d8e07179e59e',4015,'2026-03-24 05:33:59.562','pull','ready','','2026-08-19 00:57:03.968','2026-08-19 00:57:03.968',NULL);
/*!40000 ALTER TABLE `docker_images` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `docker_networks`
--

DROP TABLE IF EXISTS `docker_networks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `docker_networks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned DEFAULT NULL,
  `project_id` bigint unsigned DEFAULT NULL,
  `owner_id` bigint unsigned NOT NULL DEFAULT '0',
  `name` varchar(64) NOT NULL COMMENT '云网络名（对外）',
  `docker_name` varchar(64) NOT NULL COMMENT 'Docker 网络名 dxn-xxxxxx',
  `docker_network_id` varchar(128) NOT NULL DEFAULT '',
  `driver` varchar(32) NOT NULL DEFAULT 'bridge',
  `subnet` varchar(64) NOT NULL DEFAULT '',
  `gateway` varchar(64) NOT NULL DEFAULT '',
  `ip_range` varchar(64) NOT NULL DEFAULT '',
  `internal` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_docker_networks_name` (`name`),
  UNIQUE KEY `uk_docker_networks_docker` (`docker_name`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='云网络（底层 Docker Network）';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `docker_networks`
--

LOCK TABLES `docker_networks` WRITE;
/*!40000 ALTER TABLE `docker_networks` DISABLE KEYS */;
INSERT INTO `docker_networks` VALUES (1,NULL,NULL,1,'net-a','dxn-781421','d443ad928904a17494d4e0711ef9e73b87048d84ef721ace80a95bfcbfb1ef10','bridge','10.30.0.0/24','10.30.0.1','',0,'2026-08-19 00:49:51.452','2026-08-19 00:49:52.968','2026-08-19 00:49:52.968');
/*!40000 ALTER TABLE `docker_networks` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `docker_volumes`
--

DROP TABLE IF EXISTS `docker_volumes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `docker_volumes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned DEFAULT NULL,
  `project_id` bigint unsigned DEFAULT NULL,
  `owner_id` bigint unsigned NOT NULL DEFAULT '0',
  `name` varchar(64) NOT NULL COMMENT '云磁盘名（对外）',
  `docker_name` varchar(64) NOT NULL COMMENT 'Docker 卷名 dxv-xxxxxx',
  `driver` varchar(32) NOT NULL DEFAULT 'local',
  `mountpoint` varchar(255) NOT NULL DEFAULT '',
  `capacity_gb` int NOT NULL DEFAULT '10' COMMENT '软配额 GB',
  `used_mb` bigint NOT NULL DEFAULT '0',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_docker_volumes_name` (`name`),
  UNIQUE KEY `uk_docker_volumes_docker` (`docker_name`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='云磁盘（底层 Docker Volume）';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `docker_volumes`
--

LOCK TABLES `docker_volumes` WRITE;
/*!40000 ALTER TABLE `docker_volumes` DISABLE KEYS */;
INSERT INTO `docker_volumes` VALUES (1,NULL,NULL,1,'app-data','dxv-1641dc','local','/var/lib/docker/volumes/dxv-1641dc/_data',10,0,'2026-08-19 00:51:09.005','2026-08-19 00:51:10.738','2026-08-19 00:51:10.738'),(2,NULL,NULL,1,'app-data2','dxv-af9218','local','/var/lib/docker/volumes/dxv-af9218/_data',5,0,'2026-08-19 00:51:09.509','2026-08-19 00:51:10.766','2026-08-19 00:51:10.766');
/*!40000 ALTER TABLE `docker_volumes` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `domains`
--

DROP TABLE IF EXISTS `domains`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `domains` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned DEFAULT NULL,
  `project_id` bigint unsigned DEFAULT NULL,
  `application_id` bigint unsigned DEFAULT NULL,
  `domain` varchar(255) NOT NULL,
  `target_port` int NOT NULL DEFAULT '80',
  `tls` tinyint(1) NOT NULL DEFAULT '0',
  `cert_id` bigint unsigned DEFAULT NULL,
  `status` tinyint NOT NULL DEFAULT '1',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_domains` (`domain`),
  KEY `idx_domains_app` (`application_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `domains`
--

LOCK TABLES `domains` WRITE;
/*!40000 ALTER TABLE `domains` DISABLE KEYS */;
/*!40000 ALTER TABLE `domains` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ecs_instance_events`
--

DROP TABLE IF EXISTS `ecs_instance_events`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ecs_instance_events` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `instance_id` bigint unsigned NOT NULL,
  `event_type` varchar(32) NOT NULL,
  `level` varchar(16) NOT NULL DEFAULT 'info' COMMENT 'info/warn/error',
  `message` varchar(512) NOT NULL DEFAULT '',
  `actor_id` bigint unsigned DEFAULT NULL,
  `request_id` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_ecs_events_instance` (`instance_id`,`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=70 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='ECS 实例事件';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ecs_instance_events`
--

LOCK TABLES `ecs_instance_events` WRITE;
/*!40000 ALTER TABLE `ecs_instance_events` DISABLE KEYS */;
INSERT INTO `ecs_instance_events` VALUES (1,1,'create','info','创建实例请求已受理',1,'req-46fa334f','2026-08-19 00:13:14.145'),(2,1,'create','info','实例创建成功并已启动',1,'req-46fa334f','2026-08-19 00:13:14.322'),(3,1,'stop','info','执行 stop',1,'req-d02c48f8','2026-08-19 00:13:17.402'),(4,1,'reconcile','info','状态转换完成：running',NULL,'','2026-08-19 00:13:26.050'),(5,1,'stop','info','stop 完成，状态 stopped',1,'req-d02c48f8','2026-08-19 00:13:27.570'),(6,1,'start','info','执行 start',1,'req-a39910be','2026-08-19 00:13:27.594'),(7,1,'start','info','start 完成，状态 running',1,'req-a39910be','2026-08-19 00:13:27.714'),(8,1,'restart','info','执行 restart',1,'req-25fd9b62','2026-08-19 00:13:27.738'),(9,1,'restart','info','restart 完成，状态 running',1,'req-25fd9b62','2026-08-19 00:13:38.006'),(10,1,'force-stop','info','执行 force-stop',1,'req-62c15008','2026-08-19 00:13:38.027'),(11,1,'force-stop','info','force-stop 完成，状态 stopped',1,'req-62c15008','2026-08-19 00:13:38.173'),(12,2,'create','info','创建实例请求已受理',1,'req-54982b51','2026-08-19 00:13:54.284'),(13,2,'create','info','实例创建成功并已启动',1,'req-54982b51','2026-08-19 00:13:54.482'),(14,3,'create','info','创建实例请求已受理',2,'req-47237c45','2026-08-19 00:14:22.710'),(15,3,'create','info','实例创建成功并已启动',2,'req-47237c45','2026-08-19 00:14:22.888'),(16,4,'create','info','创建实例请求已受理',2,'req-bb222ed1','2026-08-19 00:14:22.906'),(17,4,'create','info','实例创建成功并已启动',2,'req-bb222ed1','2026-08-19 00:14:23.104'),(18,5,'create','info','创建实例请求已受理',2,'req-5d8cd52e','2026-08-19 00:14:23.120'),(19,5,'create','info','实例创建成功并已启动',2,'req-5d8cd52e','2026-08-19 00:14:23.315'),(20,6,'create','info','创建实例请求已受理',2,'req-862d1694','2026-08-19 00:14:23.342'),(21,6,'create','info','实例创建成功并已启动',2,'req-862d1694','2026-08-19 00:14:23.526'),(22,7,'create','info','创建实例请求已受理',2,'req-7d2047f8','2026-08-19 00:14:23.540'),(23,7,'create','info','实例创建成功并已启动',2,'req-7d2047f8','2026-08-19 00:14:23.722'),(24,3,'reconcile','error','容器消失，状态置 Unknown',NULL,'','2026-08-19 00:14:26.057'),(25,3,'delete','info','删除实例',2,'req-152cec21','2026-08-19 00:15:08.986'),(26,1,'delete','info','删除实例',1,'req-752000a6','2026-08-19 00:15:19.944'),(27,2,'delete','info','删除实例',1,'req-74c3d182','2026-08-19 00:15:19.981'),(28,4,'delete','info','删除实例',1,'req-19fca538','2026-08-19 00:15:20.247'),(29,5,'delete','info','删除实例',1,'req-bdc3196c','2026-08-19 00:15:20.422'),(30,6,'delete','info','删除实例',1,'req-543e994f','2026-08-19 00:15:20.643'),(31,7,'delete','info','删除实例',1,'req-7fe4ec73','2026-08-19 00:15:20.887'),(32,8,'create','info','创建实例请求已受理',1,'req-9bc2cb82','2026-08-19 00:24:56.076'),(33,8,'create','info','实例创建成功并已启动',1,'req-9bc2cb82','2026-08-19 00:24:56.265'),(34,8,'delete','info','删除实例',1,'req-9efc5dad','2026-08-19 00:36:07.403'),(35,9,'create','info','创建实例请求已受理',1,'req-c438ea8c','2026-08-19 00:48:32.708'),(36,9,'create','info','实例创建成功并已启动',1,'req-c438ea8c','2026-08-19 00:48:32.931'),(37,10,'create','info','创建实例请求已受理',1,'req-e6ec64d9','2026-08-19 00:48:35.096'),(38,10,'create','info','实例创建成功并已启动',1,'req-e6ec64d9','2026-08-19 00:48:35.318'),(39,9,'delete','info','删除实例',1,'req-44b8e88e','2026-08-19 00:48:35.420'),(40,10,'delete','info','删除实例',1,'req-528c1fb8','2026-08-19 00:48:35.645'),(41,11,'create','info','创建实例请求已受理',1,'req-36a52074','2026-08-19 00:49:51.492'),(42,11,'create','info','实例创建成功并已启动',1,'req-36a52074','2026-08-19 00:49:51.762'),(43,12,'create','info','创建实例请求已受理',1,'req-b97bc934','2026-08-19 00:49:51.876'),(44,12,'create','info','实例创建成功并已启动',1,'req-b97bc934','2026-08-19 00:49:52.101'),(45,11,'delete','info','删除实例',1,'req-23e17b2b','2026-08-19 00:49:52.346'),(46,12,'delete','info','删除实例',1,'req-bad5531e','2026-08-19 00:49:52.523'),(47,13,'create','info','创建实例请求已受理',1,'req-e363b667','2026-08-19 00:51:09.045'),(48,13,'create','info','实例创建成功并已启动',1,'req-e363b667','2026-08-19 00:51:09.318'),(49,13,'volume.attach','info','挂载云磁盘 app-data2 → /logs',1,'req-8f9ea17d','2026-08-19 00:51:09.907'),(50,13,'volume.detach','info','卸载云磁盘 app-data2',1,'req-d6feb197','2026-08-19 00:51:10.489'),(51,13,'delete','info','删除实例',1,'req-76570e98','2026-08-19 00:51:10.544'),(52,14,'create','info','创建实例请求已受理',1,'req-f16dbad8','2026-08-19 02:05:40.708'),(53,14,'create','info','实例创建成功并已启动',1,'req-f16dbad8','2026-08-19 02:05:40.897'),(54,14,'stop','info','执行 stop',1,'req-d66dcec8','2026-08-19 02:08:11.380'),(55,14,'reconcile','info','状态转换完成：running',NULL,'','2026-08-19 02:08:13.053'),(56,14,'stop','info','stop 完成，状态 stopped',1,'req-d66dcec8','2026-08-19 02:08:21.532'),(57,14,'start','info','执行 start',1,'req-f8895abe','2026-08-19 02:08:21.553'),(58,14,'start','info','start 完成，状态 running',1,'req-f8895abe','2026-08-19 02:08:21.677'),(59,14,'delete','info','删除实例',1,'req-8ede1d15','2026-08-19 02:08:34.181'),(60,15,'create','info','创建实例请求已受理',1,'req-a0206591','2026-08-19 02:29:51.389'),(61,15,'create','info','实例创建成功并已启动',1,'req-a0206591','2026-08-19 02:29:51.761'),(62,16,'create','info','创建实例请求已受理',1,'req-a26d3d1d','2026-08-19 02:29:55.206'),(63,16,'create','info','实例创建成功并已启动',1,'req-a26d3d1d','2026-08-19 02:29:55.470'),(64,17,'create','info','创建实例请求已受理',1,'req-035c0967','2026-08-19 02:33:11.295'),(65,17,'create','info','实例创建成功并已启动',1,'req-035c0967','2026-08-19 02:33:11.529'),(66,18,'create','info','创建实例请求已受理',1,'req-97f0d3cc','2026-08-19 02:33:14.844'),(67,18,'create','info','实例创建成功并已启动',1,'req-97f0d3cc','2026-08-19 02:33:15.052'),(68,19,'create','info','创建实例请求已受理',1,'req-50ec209d','2026-08-19 03:07:37.824'),(69,19,'create','info','实例创建成功并已启动',1,'req-50ec209d','2026-08-19 03:07:38.030');
/*!40000 ALTER TABLE `ecs_instance_events` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ecs_instances`
--

DROP TABLE IF EXISTS `ecs_instances`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `ecs_instances` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `instance_no` varchar(32) NOT NULL COMMENT '对外实例 ID，形如 i-a1b2c3d4e5f60718',
  `org_id` bigint unsigned DEFAULT NULL,
  `project_id` bigint unsigned DEFAULT NULL,
  `owner_id` bigint unsigned NOT NULL,
  `name` varchar(64) NOT NULL,
  `description` varchar(255) NOT NULL DEFAULT '',
  `image` varchar(255) NOT NULL,
  `cpu` decimal(6,2) NOT NULL DEFAULT '1.00' COMMENT '核数',
  `memory_mb` int NOT NULL DEFAULT '512',
  `disk_gb` int NOT NULL DEFAULT '10' COMMENT '逻辑磁盘配额 GB',
  `network_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'Docker 网络 ID（Phase 5 接入）',
  `fixed_ip` varchar(64) NOT NULL DEFAULT '',
  `ports` text COMMENT '端口映射 JSON',
  `env` text COMMENT '环境变量 JSON 数组',
  `command` text COMMENT '启动命令 JSON 数组',
  `mounts` text COMMENT '挂载 JSON（Phase 5）',
  `restart_policy` varchar(32) NOT NULL DEFAULT 'no',
  `readonly_rootfs` tinyint(1) NOT NULL DEFAULT '0',
  `desired_state` varchar(16) NOT NULL DEFAULT 'creating',
  `observed_state` varchar(16) NOT NULL DEFAULT 'creating',
  `container_id` varchar(128) DEFAULT NULL,
  `container_name` varchar(128) NOT NULL DEFAULT '',
  `last_error` text,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ecs_instance_no` (`instance_no`),
  UNIQUE KEY `uk_ecs_container_id` (`container_id`),
  KEY `idx_ecs_owner` (`owner_id`),
  KEY `idx_ecs_org_proj` (`org_id`,`project_id`),
  KEY `idx_ecs_state` (`desired_state`,`observed_state`),
  KEY `idx_ecs_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=20 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='ECS 云主机（底层 Docker 容器）';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ecs_instances`
--

LOCK TABLES `ecs_instances` WRITE;
/*!40000 ALTER TABLE `ecs_instances` DISABLE KEYS */;
INSERT INTO `ecs_instances` VALUES (1,'i-0d9adf92cf6c6861',NULL,NULL,1,'test-web','','alpine:3.20',0.50,256,10,'','','[]','null','[\"sh\",\"-c\",\"echo hello-dxcloud \\u0026\\u0026 sleep 3600\"]','','no',0,'deleting','stopped','c34f58bec822301eb52c27f978ac1d205de30b9ad7b3213668cdf62ca67869f2','dx-i-0d9adf92cf6c6861','','2026-08-19 00:13:14.141','2026-08-19 00:15:19.961','2026-08-19 00:15:19.961'),(2,'i-091d68b04e97d6ad',NULL,NULL,1,'port-a','','alpine:3.20',1.00,512,10,'','','[{\"container_port\":8080,\"host_port\":18080,\"protocol\":\"tcp\"}]','null','[\"sleep\",\"3600\"]','','no',0,'deleting','running','9eb90aa7e12c474d37e688577e989fadba6268a005dbfb7c106a4998c0667a11','dx-i-091d68b04e97d6ad','','2026-08-19 00:13:54.280','2026-08-19 00:15:20.229','2026-08-19 00:15:20.229'),(3,'i-0bcb08fc56dbbb76',NULL,NULL,2,'q1','','alpine:3.20',0.25,128,10,'','','[]','null','[\"sleep\",\"3600\"]','','no',0,'deleting','unknown','b7a57921aa9e05919c04952d3376b32f86b1eb884eb00901a56d55b970f73d8d','dx-i-0bcb08fc56dbbb76','container missing (removed outside platform?)','2026-08-19 00:14:22.708','2026-08-19 00:15:08.992','2026-08-19 00:15:08.992'),(4,'i-f2f2188374bda892',NULL,NULL,2,'q2','','alpine:3.20',0.25,128,10,'','','[]','null','[\"sleep\",\"3600\"]','','no',0,'deleting','running','d18f0e42997b1e9ef9a57746e29840d7519c3f2c0dc36153ea6a5de479306697','dx-i-f2f2188374bda892','','2026-08-19 00:14:22.902','2026-08-19 00:15:20.403','2026-08-19 00:15:20.403'),(5,'i-07accab39b9b5872',NULL,NULL,2,'q3','','alpine:3.20',0.25,128,10,'','','[]','null','[\"sleep\",\"3600\"]','','no',0,'deleting','running','7c79f9776fae98e5b367f9b23309641cd22f05aa1259b29e16e052da4a42dd42','dx-i-07accab39b9b5872','','2026-08-19 00:14:23.115','2026-08-19 00:15:20.625','2026-08-19 00:15:20.625'),(6,'i-b4633bb999715b99',NULL,NULL,2,'q4','','alpine:3.20',0.25,128,10,'','','[]','null','[\"sleep\",\"3600\"]','','no',0,'deleting','running','7137a05c9d2982487088d50846b0fb8fdfbc67dd2733235e582eb84af95884e6','dx-i-b4633bb999715b99','','2026-08-19 00:14:23.338','2026-08-19 00:15:20.855','2026-08-19 00:15:20.855'),(7,'i-d3d78d7475201067',NULL,NULL,2,'q5','','alpine:3.20',0.25,128,10,'','','[]','null','[\"sleep\",\"3600\"]','','no',0,'deleting','running','610561456bb7362ed43959707cba9a5632c6fdd4307a246d90a614be447f9cd8','dx-i-d3d78d7475201067','','2026-08-19 00:14:23.536','2026-08-19 00:15:21.095','2026-08-19 00:15:21.095'),(8,'i-f5b41e50f40f57cb',NULL,NULL,1,'term-test','','alpine:3.20',1.00,512,10,'','','[]','null','[\"sleep\",\"3600\"]','','no',0,'deleting','running','3701c91e8e391de4decfef6512c37f5366b639ef9600dfac57488928b4e85c24','dx-i-f5b41e50f40f57cb','','2026-08-19 00:24:56.072','2026-08-19 00:36:07.594','2026-08-19 00:36:07.594'),(9,'i-edadff776dd52cd8',NULL,NULL,1,'net-web','','alpine:3.20',1.00,512,10,'','','[]','null','[\"sleep\",\"3600\"]','[]','no',0,'deleting','running','a233bcd11a4e12f2e8a48afca1cb5065686d958c3ca1428dbb0babc1d743302f','dx-i-edadff776dd52cd8','','2026-08-19 00:48:32.703','2026-08-19 00:48:35.621','2026-08-19 00:48:35.621'),(10,'i-9d21eed2d323b228',NULL,NULL,1,'net-db','','alpine:3.20',1.00,512,10,'','','[]','null','[\"sleep\",\"3600\"]','[]','no',0,'deleting','running','27fe45f4bed56c2fa92368c88b91c1d84cb0291eadaab2ffe705608336c36c15','dx-i-9d21eed2d323b228','','2026-08-19 00:48:35.091','2026-08-19 00:48:35.817','2026-08-19 00:48:35.817'),(11,'i-0c924f733cb9c034',NULL,NULL,1,'net-web','','alpine:3.20',1.00,512,10,'d443ad928904a17494d4e0711ef9e73b87048d84ef721ace80a95bfcbfb1ef10','','[]','null','[\"sleep\",\"3600\"]','[]','no',0,'deleting','running','2bc305fc44b8d611dd1a7320f2a8b38c83eb3a593c997e8c5d7241798d8fdfd9','dx-i-0c924f733cb9c034','','2026-08-19 00:49:51.488','2026-08-19 00:49:52.498','2026-08-19 00:49:52.498'),(12,'i-09f5134845cb9ae9',NULL,NULL,1,'net-db','','alpine:3.20',1.00,512,10,'','','[]','null','[\"sleep\",\"3600\"]','[]','no',0,'deleting','running','6c1c3832c2ae9bb21d36565aacf980adbf8463965be39390d3f251af831c216f','dx-i-09f5134845cb9ae9','','2026-08-19 00:49:51.871','2026-08-19 00:49:52.681','2026-08-19 00:49:52.681'),(13,'i-9beedd088196d217',NULL,NULL,1,'vol-web','','alpine:3.20',1.00,512,10,'','172.17.0.2','[]','null','[\"sleep\",\"3600\"]','[{\"volume_name\":\"dxv-1641dc\",\"target\":\"/data\",\"read_only\":false}]','no',0,'deleting','running','42e9af29f835564b01dc3768b6f4492e61808d58b9689f73ec1d7c4e996857fb','dx-i-9beedd088196d217','','2026-08-19 00:51:09.040','2026-08-19 00:51:10.713','2026-08-19 00:51:10.713'),(14,'i-673c58dbdabe0455',NULL,NULL,1,'mon-demo','','alpine:3.20',1.00,512,10,'','172.17.0.2','[]','null','[\"sleep\",\"3600\"]','[]','no',0,'deleting','running','6ab71246c3f774eb6b5c8fa28b91d81090b793f1591afea08709c22b03d1500b','dx-i-673c58dbdabe0455','','2026-08-19 02:05:40.703','2026-08-19 02:08:34.354','2026-08-19 02:08:34.354'),(15,'i-8868a5136c00b7d3',1,NULL,1,'tenant-a-ecs1','','busybox:latest',1.00,512,10,'','172.17.0.2','[]','null','[\"sleep\",\"3600\"]','[]','no',0,'running','running','799a08a3d20209c6168c63a4bad70b05cc33e6d802d5d7f2c8193c771908b313','dx-i-8868a5136c00b7d3','','2026-08-19 02:29:51.383','2026-08-19 02:29:51.754',NULL),(16,'i-f82fbf823e70dee8',1,NULL,1,'tenant-a-ecs3','','busybox:latest',1.00,512,10,'','172.17.0.3','[]','null','[\"sleep\",\"3600\"]','[]','no',0,'running','running','02dbc66454cc2f2d942651e2d390520654ad04dab0ae99193275c31288f9dc7d','dx-i-f82fbf823e70dee8','','2026-08-19 02:29:55.201','2026-08-19 02:29:55.462',NULL),(17,'i-79258b28f6e449cf',3,NULL,1,'tenant-a-ecs1','','busybox:latest',1.00,512,10,'','172.17.0.4','[]','null','[\"sleep\",\"3600\"]','[]','no',0,'running','running','a2da7a902eeedcd9d34e8b71e56ae01be225d870ccd82962176b66b5c3361f55','dx-i-79258b28f6e449cf','','2026-08-19 02:33:11.292','2026-08-19 02:33:11.524',NULL),(18,'i-f2769eac89fcc103',3,NULL,1,'tenant-a-ecs3','','busybox:latest',1.00,512,10,'','172.17.0.5','[]','null','[\"sleep\",\"3600\"]','[]','no',0,'running','running','272cbf86918752856f948e264eed0d34854e4ec603967f95daf615437a152ad1','dx-i-f2769eac89fcc103','','2026-08-19 02:33:14.841','2026-08-19 02:33:15.047',NULL),(19,'i-3c155604e70975b7',NULL,NULL,1,'prod-smoke-ecs','','busybox:latest',1.00,256,10,'','172.17.0.6','[]','null','[\"sleep\",\"3600\"]','[]','no',0,'running','running','1688d57890e8ff4ff521a2e808da2d84826a1d47c5ac21b154f5600e3dc4041e','dx-i-3c155604e70975b7','','2026-08-19 03:07:37.819','2026-08-19 03:07:38.027',NULL);
/*!40000 ALTER TABLE `ecs_instances` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `login_logs`
--

DROP TABLE IF EXISTS `login_logs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `login_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned DEFAULT NULL,
  `ip` varchar(64) NOT NULL DEFAULT '',
  `user_agent` varchar(512) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '1=成功 2=失败',
  `message` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_login_logs_user` (`user_id`,`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=90 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='登录日志';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `login_logs`
--

LOCK TABLES `login_logs` WRITE;
/*!40000 ALTER TABLE `login_logs` DISABLE KEYS */;
INSERT INTO `login_logs` VALUES (1,1,'172.21.0.1','curl/8.17.0',1,'success','2026-08-18 23:46:51.689'),(2,1,'172.21.0.1','curl/8.17.0',2,'wrong password','2026-08-18 23:46:51.766'),(3,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-18 23:47:40.258'),(4,2,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-18 23:49:13.896'),(5,2,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-18 23:52:09.775'),(6,2,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-18 23:55:36.835'),(7,2,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-18 23:55:36.945'),(8,2,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-18 23:55:52.774'),(9,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-18 23:55:52.886'),(10,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:13:14.095'),(11,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:13:54.182'),(12,2,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:13:54.682'),(13,2,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:14:22.687'),(14,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:15:19.927'),(15,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:24:56.037'),(16,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:27:36.555'),(17,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:29:09.837'),(18,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:32:13.592'),(19,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:32:35.877'),(20,1,'172.21.0.1','node',1,'success','2026-08-19 00:35:29.067'),(21,1,'172.21.0.1','node',1,'success','2026-08-19 00:35:49.022'),(22,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:36:06.950'),(23,2,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:36:07.052'),(24,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:48:12.617'),(25,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:48:32.234'),(26,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:49:51.366'),(27,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:51:08.937'),(28,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:51:21.698'),(29,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:53:06.299'),(30,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:55:00.536'),(31,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 00:57:02.849'),(32,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:06:58.923'),(33,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:11:33.103'),(34,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:13:50.913'),(35,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:15:51.111'),(36,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:17:10.333'),(37,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:24:32.797'),(38,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:24:52.989'),(39,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:26:30.691'),(40,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:28:00.392'),(41,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:34:17.952'),(42,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:34:28.618'),(43,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:34:44.741'),(44,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:36:15.747'),(45,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:37:36.047'),(46,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:37:47.608'),(47,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:39:03.104'),(48,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:39:14.679'),(49,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:40:19.031'),(50,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:41:16.106'),(51,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:44:11.843'),(52,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:46:06.630'),(53,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:46:42.402'),(54,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:47:03.146'),(55,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:48:03.631'),(56,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:49:46.338'),(57,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:50:10.056'),(58,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 01:50:21.129'),(59,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:05:40.655'),(60,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:08:11.341'),(61,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:08:34.146'),(62,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:28:45.861'),(63,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:29:50.632'),(64,3,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:29:51.057'),(65,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:30:36.779'),(66,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:31:56.737'),(67,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:32:06.009'),(68,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:32:12.952'),(69,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:33:10.833'),(70,4,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:33:11.082'),(71,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:36:32.184'),(72,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:43:14.346'),(73,5,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',2,'wrong password','2026-08-19 02:43:14.964'),(74,5,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',2,'wrong password','2026-08-19 02:43:15.036'),(75,5,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',2,'wrong password','2026-08-19 02:43:15.104'),(76,5,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',2,'wrong password','2026-08-19 02:43:15.170'),(77,5,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:44:17.513'),(78,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:44:31.752'),(79,6,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',2,'wrong password','2026-08-19 02:44:32.311'),(80,6,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',2,'wrong password','2026-08-19 02:44:32.383'),(81,6,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',2,'wrong password','2026-08-19 02:44:32.454'),(82,6,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:45:34.802'),(83,1,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:45:48.359'),(84,7,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',2,'wrong password','2026-08-19 02:46:53.941'),(85,7,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',2,'wrong password','2026-08-19 02:46:54.009'),(86,7,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',2,'wrong password','2026-08-19 02:46:54.076'),(87,7,'172.21.0.1','Mozilla/5.0 (Windows NT; Windows NT 10.0; zh-CN) WindowsPowerShell/5.1.26100.7670',1,'success','2026-08-19 02:47:56.457'),(88,1,'172.21.0.1','curl/8.17.0',1,'success','2026-08-19 03:06:49.496'),(89,1,'172.21.0.1','curl/8.17.0',1,'success','2026-08-19 03:07:37.726');
/*!40000 ALTER TABLE `login_logs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `metric_samples`
--

DROP TABLE IF EXISTS `metric_samples`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `metric_samples` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `kind` varchar(8) NOT NULL DEFAULT 'ecs' COMMENT 'ecs/app',
  `ref_id` bigint unsigned NOT NULL COMMENT '实例/部署 ID',
  `ts` datetime(3) NOT NULL,
  `cpu_pct` decimal(8,2) NOT NULL DEFAULT '0.00',
  `mem_used` bigint NOT NULL DEFAULT '0',
  `mem_limit` bigint NOT NULL DEFAULT '0',
  `net_rx` bigint NOT NULL DEFAULT '0',
  `net_tx` bigint NOT NULL DEFAULT '0',
  `disk_read` bigint NOT NULL DEFAULT '0',
  `disk_write` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `idx_metric_ref_ts` (`kind`,`ref_id`,`ts`),
  KEY `idx_metric_ts` (`ts`)
) ENGINE=InnoDB AUTO_INCREMENT=89 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='监控指标采样（分钟级）';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `metric_samples`
--

LOCK TABLES `metric_samples` WRITE;
/*!40000 ALTER TABLE `metric_samples` DISABLE KEYS */;
INSERT INTO `metric_samples` VALUES (1,'ecs',14,'2026-08-19 02:06:33.747',0.00,1052672,536870912,1172,126,0,0),(2,'ecs',14,'2026-08-19 02:08:31.058',0.00,352256,536870912,1214,126,0,0),(3,'ecs',15,'2026-08-19 02:29:59.694',0.00,1392640,536870912,1298,126,0,0),(4,'ecs',16,'2026-08-19 02:30:00.700',0.00,1748992,536870912,872,126,0,0),(5,'ecs',16,'2026-08-19 02:30:59.608',0.00,1245184,536870912,872,126,0,0),(6,'ecs',15,'2026-08-19 02:31:00.610',0.00,1241088,536870912,1298,126,0,0),(7,'ecs',16,'2026-08-19 02:32:47.450',0.00,450560,536870912,872,126,0,0),(8,'ecs',15,'2026-08-19 02:32:47.455',0.00,446464,536870912,1298,126,0,0),(9,'ecs',16,'2026-08-19 02:33:47.446',0.00,450560,536870912,1124,126,0,0),(10,'ecs',18,'2026-08-19 02:33:47.451',0.00,1257472,536870912,872,126,0,0),(11,'ecs',17,'2026-08-19 02:33:47.453',0.00,3399680,536870912,998,126,0,0),(12,'ecs',15,'2026-08-19 02:33:47.456',0.00,446464,536870912,1550,126,0,0),(13,'ecs',18,'2026-08-19 02:35:36.052',0.00,450560,536870912,872,126,0,0),(14,'ecs',16,'2026-08-19 02:35:36.061',0.00,450560,536870912,1124,126,0,0),(15,'ecs',17,'2026-08-19 02:35:37.076',0.00,581632,536870912,998,126,0,0),(16,'ecs',15,'2026-08-19 02:35:37.077',0.00,446464,536870912,1550,126,0,0),(17,'ecs',15,'2026-08-19 02:36:36.045',0.00,446464,536870912,1550,126,0,0),(18,'ecs',17,'2026-08-19 02:36:37.060',0.00,581632,536870912,998,126,0,0),(19,'ecs',18,'2026-08-19 02:36:37.061',0.00,450560,536870912,872,126,0,0),(20,'ecs',16,'2026-08-19 02:36:37.067',0.00,450560,536870912,1124,126,0,0),(21,'ecs',17,'2026-08-19 02:37:36.041',0.00,581632,536870912,998,126,0,0),(22,'ecs',18,'2026-08-19 02:37:37.058',0.00,450560,536870912,872,126,0,0),(23,'ecs',16,'2026-08-19 02:37:37.063',0.00,450560,536870912,1124,126,0,0),(24,'ecs',15,'2026-08-19 02:37:37.065',0.00,446464,536870912,1550,126,0,0),(25,'ecs',15,'2026-08-19 02:38:36.044',0.00,446464,536870912,1550,126,0,0),(26,'ecs',17,'2026-08-19 02:38:37.050',0.00,581632,536870912,998,126,0,0),(27,'ecs',16,'2026-08-19 02:38:37.053',0.00,450560,536870912,1124,126,0,0),(28,'ecs',18,'2026-08-19 02:38:37.054',0.00,450560,536870912,872,126,0,0),(29,'ecs',15,'2026-08-19 02:39:36.035',0.00,446464,536870912,1550,126,0,0),(30,'ecs',17,'2026-08-19 02:39:37.046',0.00,581632,536870912,998,126,0,0),(31,'ecs',16,'2026-08-19 02:39:37.055',0.00,450560,536870912,1124,126,0,0),(32,'ecs',18,'2026-08-19 02:39:37.055',0.00,450560,536870912,872,126,0,0),(33,'ecs',16,'2026-08-19 02:40:36.034',0.00,450560,536870912,1124,126,0,0),(34,'ecs',17,'2026-08-19 02:40:37.040',0.00,581632,536870912,998,126,0,0),(35,'ecs',18,'2026-08-19 02:40:37.044',0.00,450560,536870912,872,126,0,0),(36,'ecs',15,'2026-08-19 02:40:37.045',0.00,446464,536870912,1550,126,0,0),(37,'ecs',17,'2026-08-19 02:42:37.927',0.00,581632,536870912,998,126,0,0),(38,'ecs',18,'2026-08-19 02:42:37.928',0.00,450560,536870912,872,126,0,0),(39,'ecs',15,'2026-08-19 02:42:37.933',0.00,446464,536870912,1550,126,0,0),(40,'ecs',16,'2026-08-19 02:42:37.936',0.00,450560,536870912,1124,126,0,0),(41,'ecs',15,'2026-08-19 02:43:37.919',0.00,446464,536870912,1550,126,0,0),(42,'ecs',16,'2026-08-19 02:43:37.922',0.00,450560,536870912,1124,126,0,0),(43,'ecs',17,'2026-08-19 02:43:37.925',0.00,581632,536870912,998,126,0,0),(44,'ecs',18,'2026-08-19 02:43:38.932',0.00,450560,536870912,872,126,0,0),(45,'ecs',16,'2026-08-19 02:44:37.913',0.00,450560,536870912,1124,126,0,0),(46,'ecs',15,'2026-08-19 02:44:37.919',0.00,446464,536870912,1550,126,0,0),(47,'ecs',18,'2026-08-19 02:44:37.921',0.00,450560,536870912,872,126,0,0),(48,'ecs',17,'2026-08-19 02:44:38.927',0.00,581632,536870912,998,126,0,0),(49,'ecs',15,'2026-08-19 02:45:37.913',0.00,446464,536870912,1550,126,0,0),(50,'ecs',16,'2026-08-19 02:45:37.916',0.00,450560,536870912,1124,126,0,0),(51,'ecs',18,'2026-08-19 02:45:37.922',0.00,450560,536870912,872,126,0,0),(52,'ecs',17,'2026-08-19 02:45:38.926',0.00,581632,536870912,998,126,0,0),(53,'ecs',17,'2026-08-19 02:46:37.908',0.00,581632,536870912,998,126,0,0),(54,'ecs',18,'2026-08-19 02:46:37.915',0.00,450560,536870912,872,126,0,0),(55,'ecs',15,'2026-08-19 02:46:37.916',0.00,446464,536870912,1550,126,0,0),(56,'ecs',16,'2026-08-19 02:46:37.917',0.00,450560,536870912,1124,126,0,0),(57,'ecs',16,'2026-08-19 02:47:37.909',0.00,450560,536870912,1124,126,0,0),(58,'ecs',18,'2026-08-19 02:47:37.911',0.00,450560,536870912,872,126,0,0),(59,'ecs',15,'2026-08-19 02:47:37.914',0.00,446464,536870912,1550,126,0,0),(60,'ecs',17,'2026-08-19 02:47:37.918',0.00,581632,536870912,998,126,0,0),(61,'ecs',17,'2026-08-19 02:48:37.906',0.00,581632,536870912,998,126,0,0),(62,'ecs',16,'2026-08-19 02:48:37.910',0.00,450560,536870912,1124,126,0,0),(63,'ecs',15,'2026-08-19 02:48:37.911',0.00,446464,536870912,1550,126,0,0),(64,'ecs',18,'2026-08-19 02:48:37.912',0.00,450560,536870912,872,126,0,0),(65,'ecs',18,'2026-08-19 02:50:27.189',0.00,450560,536870912,872,126,0,0),(66,'ecs',16,'2026-08-19 02:50:28.198',0.00,450560,536870912,1124,126,0,0),(67,'ecs',15,'2026-08-19 02:50:28.199',0.00,446464,536870912,1550,126,0,0),(68,'ecs',17,'2026-08-19 02:50:28.206',0.00,581632,536870912,998,126,0,0),(69,'ecs',17,'2026-08-19 02:51:28.187',0.00,581632,536870912,998,126,0,0),(70,'ecs',15,'2026-08-19 02:51:28.192',0.00,446464,536870912,1550,126,0,0),(71,'ecs',16,'2026-08-19 02:51:28.195',0.00,450560,536870912,1124,126,0,0),(72,'ecs',18,'2026-08-19 02:51:28.198',0.00,450560,536870912,872,126,0,0),(73,'ecs',18,'2026-08-19 02:52:27.179',0.00,450560,536870912,998,126,0,0),(74,'ecs',15,'2026-08-19 02:52:28.195',0.00,446464,536870912,1676,126,0,0),(75,'ecs',17,'2026-08-19 02:52:28.198',0.00,581632,536870912,1124,126,0,0),(76,'ecs',16,'2026-08-19 02:52:28.202',0.00,450560,536870912,1250,126,0,0),(77,'ecs',18,'2026-08-19 02:53:28.193',0.00,450560,536870912,3428,126,0,0),(78,'ecs',17,'2026-08-19 02:53:28.195',0.00,581632,536870912,3554,126,0,0),(79,'ecs',15,'2026-08-19 02:53:28.198',0.00,446464,536870912,4106,126,0,0),(80,'ecs',16,'2026-08-19 02:53:28.199',0.00,450560,536870912,3680,126,0,0),(81,'ecs',17,'2026-08-19 02:54:28.185',0.00,581632,536870912,3746,126,0,0),(82,'ecs',18,'2026-08-19 02:54:28.189',0.00,450560,536870912,3620,126,0,0),(83,'ecs',15,'2026-08-19 02:54:28.193',0.00,446464,536870912,4298,126,0,0),(84,'ecs',16,'2026-08-19 02:54:28.195',0.00,450560,536870912,3872,126,0,0),(85,'ecs',18,'2026-08-19 03:06:27.401',0.00,450560,536870912,3662,126,0,0),(86,'ecs',17,'2026-08-19 03:06:27.404',0.00,581632,536870912,3788,126,0,0),(87,'ecs',15,'2026-08-19 03:06:28.413',0.00,446464,536870912,4340,126,0,0),(88,'ecs',16,'2026-08-19 03:06:28.417',0.00,450560,536870912,3914,126,0,0);
/*!40000 ALTER TABLE `metric_samples` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `notifications`
--

DROP TABLE IF EXISTS `notifications`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `notifications` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `org_id` bigint unsigned DEFAULT NULL,
  `type` varchar(32) NOT NULL DEFAULT 'system' COMMENT 'system/pipeline/deploy',
  `title` varchar(128) NOT NULL,
  `content` varchar(512) NOT NULL DEFAULT '',
  `read_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_notifications_user` (`user_id`,`read_at`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='站内通知';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `notifications`
--

LOCK TABLES `notifications` WRITE;
/*!40000 ALTER TABLE `notifications` DISABLE KEYS */;
INSERT INTO `notifications` VALUES (1,1,NULL,'pipeline','Pipeline 运行成功','运行 #19（#11）状态：success','2026-08-19 02:08:26.791','2026-08-19 02:08:23.303');
/*!40000 ALTER TABLE `notifications` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `operation_logs`
--

DROP TABLE IF EXISTS `operation_logs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `operation_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned DEFAULT NULL,
  `user_id` bigint unsigned DEFAULT NULL,
  `module` varchar(32) NOT NULL DEFAULT '',
  `action` varchar(64) NOT NULL,
  `target_type` varchar(32) NOT NULL DEFAULT '',
  `target_id` varchar(64) NOT NULL DEFAULT '',
  `target_name` varchar(128) NOT NULL DEFAULT '',
  `result` tinyint NOT NULL DEFAULT '1' COMMENT '1=成功 2=失败',
  `duration_ms` bigint NOT NULL DEFAULT '0',
  `ip` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_oplogs_user` (`user_id`,`created_at`),
  KEY `idx_oplogs_module` (`module`,`created_at`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='操作日志（用户操作流水）';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `operation_logs`
--

LOCK TABLES `operation_logs` WRITE;
/*!40000 ALTER TABLE `operation_logs` DISABLE KEYS */;
INSERT INTO `operation_logs` VALUES (1,NULL,1,'ecs','stop','ecs','i-673c58dbdabe0455','mon-demo',1,10155,'172.21.0.1','2026-08-19 02:08:21.539'),(2,NULL,1,'ecs','start','ecs','i-673c58dbdabe0455','mon-demo',1,126,'172.21.0.1','2026-08-19 02:08:21.684'),(3,NULL,1,'ecs','delete','ecs','i-673c58dbdabe0455','mon-demo',1,0,'172.21.0.1','2026-08-19 02:08:34.364'),(4,NULL,1,'ecs','create','ecs','i-8868a5136c00b7d3','tenant-a-ecs1',1,0,'172.21.0.1','2026-08-19 02:29:51.402'),(5,NULL,1,'ecs','create','ecs','i-f82fbf823e70dee8','tenant-a-ecs3',1,0,'172.21.0.1','2026-08-19 02:29:55.215'),(6,NULL,1,'ecs','create','ecs','i-79258b28f6e449cf','tenant-a-ecs1',1,0,'172.21.0.1','2026-08-19 02:33:11.301'),(7,NULL,1,'ecs','create','ecs','i-f2769eac89fcc103','tenant-a-ecs3',1,0,'172.21.0.1','2026-08-19 02:33:14.851'),(8,NULL,1,'ecs','create','ecs','i-3c155604e70975b7','prod-smoke-ecs',1,0,'172.21.0.1','2026-08-19 03:07:37.828');
/*!40000 ALTER TABLE `operation_logs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `organization_members`
--

DROP TABLE IF EXISTS `organization_members`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `organization_members` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  `org_role` varchar(16) NOT NULL DEFAULT 'member' COMMENT 'owner/admin/member/viewer',
  `status` tinyint NOT NULL DEFAULT '1',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_org_members` (`org_id`,`user_id`),
  KEY `idx_org_members_user` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='组织成员';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `organization_members`
--

LOCK TABLES `organization_members` WRITE;
/*!40000 ALTER TABLE `organization_members` DISABLE KEYS */;
INSERT INTO `organization_members` VALUES (1,1,1,'owner',1,'2026-08-19 02:29:50.724','2026-08-19 02:29:50.724'),(2,2,1,'owner',1,'2026-08-19 02:29:50.800','2026-08-19 02:29:50.800'),(3,1,3,'member',1,'2026-08-19 02:29:51.084','2026-08-19 02:29:51.084'),(4,3,1,'owner',1,'2026-08-19 02:33:10.871','2026-08-19 02:33:10.871'),(5,4,1,'owner',1,'2026-08-19 02:33:10.907','2026-08-19 02:33:10.907'),(6,3,4,'member',1,'2026-08-19 02:33:11.113','2026-08-19 02:33:11.113');
/*!40000 ALTER TABLE `organization_members` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `organizations`
--

DROP TABLE IF EXISTS `organizations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `organizations` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `code` varchar(64) NOT NULL,
  `plan` varchar(16) NOT NULL DEFAULT 'free' COMMENT 'free/basic/pro',
  `credit` decimal(18,4) NOT NULL DEFAULT '0.0000' COMMENT '虚拟余额',
  `status` tinyint NOT NULL DEFAULT '1',
  `created_by` bigint unsigned DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_organizations_name` (`name`),
  UNIQUE KEY `uk_organizations_code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='组织（租户）';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `organizations`
--

LOCK TABLES `organizations` WRITE;
/*!40000 ALTER TABLE `organizations` DISABLE KEYS */;
INSERT INTO `organizations` VALUES (1,'??A-0819022950','tenant-a-0819022950','pro',999.5500,1,1,'2026-08-19 02:29:50.719','2026-08-19 02:33:14.584',NULL),(2,'??B-0819022950','tenant-b-0819022950','free',1000.0000,1,1,'2026-08-19 02:29:50.794','2026-08-19 02:29:50.794',NULL),(3,'??A-0819023310','tenant-a-0819023310','pro',1000.0000,1,1,'2026-08-19 02:33:10.867','2026-08-19 02:33:15.304',NULL),(4,'??B-0819023310','tenant-b-0819023310','free',1000.0000,1,1,'2026-08-19 02:33:10.904','2026-08-19 02:33:10.904',NULL);
/*!40000 ALTER TABLE `organizations` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `permissions`
--

DROP TABLE IF EXISTS `permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `permissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(64) NOT NULL COMMENT '如 ecs:create',
  `name` varchar(64) NOT NULL,
  `module` varchar(32) NOT NULL DEFAULT '' COMMENT 'ecs/image/network/...',
  `description` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_permissions_code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=76 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='权限点';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `permissions`
--

LOCK TABLES `permissions` WRITE;
/*!40000 ALTER TABLE `permissions` DISABLE KEYS */;
INSERT INTO `permissions` VALUES (1,'ecs:list','ecs:list','ecs','','2026-08-18 23:46:04.912','2026-08-18 23:46:04.912'),(2,'ecs:get','ecs:get','ecs','','2026-08-18 23:46:04.915','2026-08-18 23:46:04.915'),(3,'ecs:create','ecs:create','ecs','','2026-08-18 23:46:04.919','2026-08-18 23:46:04.919'),(4,'ecs:update','ecs:update','ecs','','2026-08-18 23:46:04.944','2026-08-18 23:46:04.944'),(5,'ecs:delete','ecs:delete','ecs','','2026-08-18 23:46:04.947','2026-08-18 23:46:04.947'),(6,'ecs:start','ecs:start','ecs','','2026-08-18 23:46:04.949','2026-08-18 23:46:04.949'),(7,'ecs:stop','ecs:stop','ecs','','2026-08-18 23:46:04.952','2026-08-18 23:46:04.952'),(8,'ecs:restart','ecs:restart','ecs','','2026-08-18 23:46:04.954','2026-08-18 23:46:04.954'),(9,'ecs:force-stop','ecs:force-stop','ecs','','2026-08-18 23:46:04.956','2026-08-18 23:46:04.956'),(10,'ecs:console','ecs:console','ecs','','2026-08-18 23:46:04.959','2026-08-18 23:46:04.959'),(11,'ecs:logs','ecs:logs','ecs','','2026-08-18 23:46:04.962','2026-08-18 23:46:04.962'),(12,'ecs:stats','ecs:stats','ecs','','2026-08-18 23:46:04.964','2026-08-18 23:46:04.964'),(13,'ecs:recreate','ecs:recreate','ecs','','2026-08-18 23:46:04.967','2026-08-18 23:46:04.967'),(14,'image:list','image:list','image','','2026-08-18 23:46:04.969','2026-08-18 23:46:04.969'),(15,'image:pull','image:pull','image','','2026-08-18 23:46:04.971','2026-08-18 23:46:04.971'),(16,'image:delete','image:delete','image','','2026-08-18 23:46:04.974','2026-08-18 23:46:04.974'),(17,'image:build','image:build','image','','2026-08-18 23:46:04.976','2026-08-18 23:46:04.976'),(18,'image:tag','image:tag','image','','2026-08-18 23:46:04.979','2026-08-18 23:46:04.979'),(19,'image:push','image:push','image','','2026-08-18 23:46:04.981','2026-08-18 23:46:04.981'),(20,'network:list','network:list','network','','2026-08-18 23:46:04.983','2026-08-18 23:46:04.983'),(21,'network:create','network:create','network','','2026-08-18 23:46:04.986','2026-08-18 23:46:04.986'),(22,'network:delete','network:delete','network','','2026-08-18 23:46:04.988','2026-08-18 23:46:04.988'),(23,'network:connect','network:connect','network','','2026-08-18 23:46:04.992','2026-08-18 23:46:04.992'),(24,'volume:list','volume:list','volume','','2026-08-18 23:46:04.996','2026-08-18 23:46:04.996'),(25,'volume:create','volume:create','volume','','2026-08-18 23:46:04.998','2026-08-18 23:46:04.998'),(26,'volume:delete','volume:delete','volume','','2026-08-18 23:46:05.023','2026-08-18 23:46:05.023'),(27,'volume:attach','volume:attach','volume','','2026-08-18 23:46:05.026','2026-08-18 23:46:05.026'),(28,'registry:list','registry:list','registry','','2026-08-18 23:46:05.029','2026-08-18 23:46:05.029'),(29,'registry:create','registry:create','registry','','2026-08-18 23:46:05.031','2026-08-18 23:46:05.031'),(30,'registry:delete','registry:delete','registry','','2026-08-18 23:46:05.034','2026-08-18 23:46:05.034'),(31,'registry:push','registry:push','registry','','2026-08-18 23:46:05.036','2026-08-18 23:46:05.036'),(32,'registry:pull','registry:pull','registry','','2026-08-18 23:46:05.039','2026-08-18 23:46:05.039'),(33,'app:list','app:list','app','','2026-08-18 23:46:05.041','2026-08-18 23:46:05.041'),(34,'app:create','app:create','app','','2026-08-18 23:46:05.044','2026-08-18 23:46:05.044'),(35,'app:update','app:update','app','','2026-08-18 23:46:05.046','2026-08-18 23:46:05.046'),(36,'app:delete','app:delete','app','','2026-08-18 23:46:05.049','2026-08-18 23:46:05.049'),(37,'app:deploy','app:deploy','app','','2026-08-18 23:46:05.051','2026-08-18 23:46:05.051'),(38,'app:rollback','app:rollback','app','','2026-08-18 23:46:05.054','2026-08-18 23:46:05.054'),(39,'pipeline:list','pipeline:list','pipeline','','2026-08-18 23:46:05.056','2026-08-18 23:46:05.056'),(40,'pipeline:create','pipeline:create','pipeline','','2026-08-18 23:46:05.059','2026-08-18 23:46:05.059'),(41,'pipeline:update','pipeline:update','pipeline','','2026-08-18 23:46:05.062','2026-08-18 23:46:05.062'),(42,'pipeline:delete','pipeline:delete','pipeline','','2026-08-18 23:46:05.065','2026-08-18 23:46:05.065'),(43,'pipeline:run','pipeline:run','pipeline','','2026-08-18 23:46:05.067','2026-08-18 23:46:05.067'),(44,'pipeline:cancel','pipeline:cancel','pipeline','','2026-08-18 23:46:05.071','2026-08-18 23:46:05.071'),(45,'project:list','project:list','project','','2026-08-18 23:46:05.073','2026-08-18 23:46:05.073'),(46,'project:create','project:create','project','','2026-08-18 23:46:05.075','2026-08-18 23:46:05.075'),(47,'project:update','project:update','project','','2026-08-18 23:46:05.099','2026-08-18 23:46:05.099'),(48,'project:delete','project:delete','project','','2026-08-18 23:46:05.102','2026-08-18 23:46:05.102'),(49,'project:deploy','project:deploy','project','','2026-08-18 23:46:05.104','2026-08-18 23:46:05.104'),(50,'domain:list','domain:list','domain','','2026-08-18 23:46:05.107','2026-08-18 23:46:05.107'),(51,'domain:create','domain:create','domain','','2026-08-18 23:46:05.109','2026-08-18 23:46:05.109'),(52,'domain:delete','domain:delete','domain','','2026-08-18 23:46:05.111','2026-08-18 23:46:05.111'),(53,'domain:bind','domain:bind','domain','','2026-08-18 23:46:05.114','2026-08-18 23:46:05.114'),(54,'user:list','user:list','user','','2026-08-18 23:46:05.116','2026-08-18 23:46:05.116'),(55,'user:create','user:create','user','','2026-08-18 23:46:05.119','2026-08-18 23:46:05.119'),(56,'user:update','user:update','user','','2026-08-18 23:46:05.123','2026-08-18 23:46:05.123'),(57,'user:delete','user:delete','user','','2026-08-18 23:46:05.127','2026-08-18 23:46:05.127'),(58,'user:grant','user:grant','user','','2026-08-18 23:46:05.130','2026-08-18 23:46:05.130'),(59,'org:list','org:list','org','','2026-08-18 23:46:05.132','2026-08-18 23:46:05.132'),(60,'org:create','org:create','org','','2026-08-18 23:46:05.135','2026-08-18 23:46:05.135'),(61,'org:update','org:update','org','','2026-08-18 23:46:05.138','2026-08-18 23:46:05.138'),(62,'org:delete','org:delete','org','','2026-08-18 23:46:05.141','2026-08-18 23:46:05.141'),(63,'org:member:manage','org:member:manage','org','','2026-08-18 23:46:05.144','2026-08-18 23:46:05.144'),(64,'quota:view','quota:view','quota','','2026-08-18 23:46:05.146','2026-08-18 23:46:05.146'),(65,'quota:update','quota:update','quota','','2026-08-18 23:46:05.148','2026-08-18 23:46:05.148'),(66,'billing:view','billing:view','billing','','2026-08-18 23:46:05.172','2026-08-18 23:46:05.172'),(67,'audit:view','audit:view','audit','','2026-08-18 23:46:05.177','2026-08-18 23:46:05.177'),(68,'settings:view','settings:view','settings','','2026-08-18 23:46:05.180','2026-08-18 23:46:05.180'),(69,'settings:update','settings:update','settings','','2026-08-18 23:46:05.184','2026-08-18 23:46:05.184'),(70,'security:view','security:view','security','','2026-08-19 02:41:33.653','2026-08-19 02:41:33.653'),(71,'security:scan','security:scan','security','','2026-08-19 02:41:33.657','2026-08-19 02:41:33.657'),(72,'secret:list','secret:list','secret','','2026-08-19 02:41:33.660','2026-08-19 02:41:33.660'),(73,'secret:create','secret:create','secret','','2026-08-19 02:41:33.664','2026-08-19 02:41:33.664'),(74,'secret:delete','secret:delete','secret','','2026-08-19 02:41:33.667','2026-08-19 02:41:33.667'),(75,'secret:reveal','secret:reveal','secret','','2026-08-19 02:41:33.670','2026-08-19 02:41:33.670');
/*!40000 ALTER TABLE `permissions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `pipeline_job_runs`
--

DROP TABLE IF EXISTS `pipeline_job_runs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pipeline_job_runs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `pipeline_run_id` bigint unsigned NOT NULL,
  `step_id` bigint unsigned NOT NULL DEFAULT '0',
  `name` varchar(64) NOT NULL,
  `type` varchar(32) NOT NULL,
  `status` varchar(16) NOT NULL DEFAULT 'pending' COMMENT 'pending/running/success/failed/skipped',
  `exit_code` int NOT NULL DEFAULT '0',
  `container_id` varchar(128) NOT NULL DEFAULT '',
  `log_path` varchar(255) NOT NULL DEFAULT '',
  `started_at` datetime(3) DEFAULT NULL,
  `finished_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_pipeline_jobs_run` (`pipeline_run_id`)
) ENGINE=InnoDB AUTO_INCREMENT=60 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Pipeline 步骤任务';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `pipeline_job_runs`
--

LOCK TABLES `pipeline_job_runs` WRITE;
/*!40000 ALTER TABLE `pipeline_job_runs` DISABLE KEYS */;
INSERT INTO `pipeline_job_runs` VALUES (1,1,1,'greet','shell','success',0,'89137f897225db51b2c0305228704cfd383b6796c1a04f61dd5e80a1b75b5006','/tmp/dxlogs/1787073872860637333.log','2026-08-19 01:24:32.881','2026-08-19 01:24:33.203','2026-08-19 01:24:32.861'),(2,1,2,'tolerate-fail','shell','skipped',3,'e6f48141467a8bca905aa834f5d0ee217063cc343de4df6a97ef3ca2f7d49219','/tmp/dxlogs/1787073872864444291.log','2026-08-19 01:24:33.207','2026-08-19 01:24:33.556','2026-08-19 01:24:32.865'),(3,1,3,'final','shell','success',0,'5b990c826007037f53cddeabe7be883838a168bca1db6876e7d56da46da70cb5','/tmp/dxlogs/1787073872867136643.log','2026-08-19 01:24:33.561','2026-08-19 01:24:33.907','2026-08-19 01:24:32.867'),(4,2,4,'boom','shell','failed',7,'404821a33db10efe98c762ec0c53d89cb3ca119bf754012d15e9e52f6a74a921','/tmp/dxlogs/1787073893057566855.log','2026-08-19 01:24:53.076','2026-08-19 01:24:53.457','2026-08-19 01:24:53.058'),(5,2,5,'never','shell','pending',0,'','/tmp/dxlogs/1787073893060077276.log',NULL,NULL,'2026-08-19 01:24:53.060'),(6,3,6,'long-sleep','shell','failed',137,'e55dafb9ff225efab6f364bbad0c935ba1ae7d6f0e534949290ea80ec30c39c9','/tmp/dxlogs/1787073898118795538.log','2026-08-19 01:24:58.133','2026-08-19 01:25:06.275','2026-08-19 01:24:58.119'),(7,3,7,'after','shell','pending',0,'','/tmp/dxlogs/1787073898121114533.log',NULL,NULL,'2026-08-19 01:24:58.121'),(8,4,8,'long-sleep','shell','failed',137,'550e7db2655a835273904c5f3cc2ab8fbff381d0eda9e5b5ecab279bf4d696c0','/tmp/dxlogs/1787073990734937088.log','2026-08-19 01:26:30.752','2026-08-19 01:26:38.896','2026-08-19 01:26:30.736'),(9,4,9,'after','shell','skipped',0,'','/tmp/dxlogs/1787073990739221414.log',NULL,'2026-08-19 01:26:38.902','2026-08-19 01:26:30.739'),(10,5,10,'boom','shell','failed',7,'a0a5e3a5ba5f3b992526b2b5f90c45566fbca6cbd508f1dc6c89a3bcc1d10cf6','/tmp/dxlogs/1787074003911758833.log','2026-08-19 01:26:43.924','2026-08-19 01:26:44.246','2026-08-19 01:26:43.912'),(11,5,11,'never','shell','skipped',0,'','/tmp/dxlogs/1787074003914122473.log',NULL,'2026-08-19 01:26:44.250','2026-08-19 01:26:43.914'),(12,6,12,'long-sleep','shell','failed',137,'7bc93d223c1a9106a54640292681713f7392c3dbf94b0a03667f21c61ac09836','/tmp/dxlogs/1787074080433117737.log','2026-08-19 01:28:00.450','2026-08-19 01:28:08.611','2026-08-19 01:28:00.433'),(13,6,13,'after','shell','skipped',0,'','/tmp/dxlogs/1787074080436076064.log',NULL,'2026-08-19 01:28:08.615','2026-08-19 01:28:00.436'),(14,7,14,'prepare','shell','success',0,'f28071c44344fb9b2eb311fa8548688456e980573d8377706b148044cca9c055','/tmp/dxlogs/1787074468666984661.log','2026-08-19 01:34:28.686','2026-08-19 01:34:29.046','2026-08-19 01:34:28.667'),(15,7,15,'build','docker-build','failed',-1,'','/tmp/dxlogs/1787074468669562640.log','2026-08-19 01:34:29.050','2026-08-19 01:34:29.054','2026-08-19 01:34:28.670'),(16,7,16,'push','docker-push','skipped',0,'','/tmp/dxlogs/1787074468672046326.log',NULL,'2026-08-19 01:34:29.402','2026-08-19 01:34:28.672'),(17,7,17,'deploy','docker-deploy','skipped',0,'','/tmp/dxlogs/1787074468674428189.log',NULL,'2026-08-19 01:34:29.408','2026-08-19 01:34:28.675'),(18,8,18,'prepare','shell','success',0,'1646b98a08af6944edcf5215165eda45b86ea0983d932c080825837c36a5062f','/tmp/dxlogs/1787074575796568787.log','2026-08-19 01:36:15.817','2026-08-19 01:36:16.157','2026-08-19 01:36:15.797'),(19,8,19,'build','docker-build','failed',-1,'','/tmp/dxlogs/1787074575799322089.log','2026-08-19 01:36:16.161','2026-08-19 01:36:16.166','2026-08-19 01:36:15.799'),(20,8,20,'push','docker-push','skipped',0,'','/tmp/dxlogs/1787074575801964583.log',NULL,'2026-08-19 01:36:16.509','2026-08-19 01:36:15.802'),(21,8,21,'deploy','docker-deploy','skipped',0,'','/tmp/dxlogs/1787074575805102926.log',NULL,'2026-08-19 01:36:16.512','2026-08-19 01:36:15.805'),(22,9,22,'prepare','shell','success',0,'79c595996d03a44d29af0d0b299f89edde5651167aac7b9bb0cd9c1033a9adbd','/tmp/dxlogs/1787074656093599857.log','2026-08-19 01:37:36.114','2026-08-19 01:37:36.507','2026-08-19 01:37:36.094'),(23,9,23,'build','docker-build','failed',-1,'','/tmp/dxlogs/1787074656096333831.log','2026-08-19 01:37:36.512','2026-08-19 01:37:36.516','2026-08-19 01:37:36.096'),(24,9,24,'push','docker-push','skipped',0,'','/tmp/dxlogs/1787074656098433864.log',NULL,'2026-08-19 01:37:36.873','2026-08-19 01:37:36.099'),(25,9,25,'deploy','docker-deploy','skipped',0,'','/tmp/dxlogs/1787074656100750509.log',NULL,'2026-08-19 01:37:36.878','2026-08-19 01:37:36.101'),(26,10,26,'prepare','shell','success',0,'1c42913d281e41a3edede8482aa2562ef3028ef69001439e2c57b3afc428d564','/tmp/dxlogs/1787074743146539348.log','2026-08-19 01:39:03.166','2026-08-19 01:39:03.523','2026-08-19 01:39:03.147'),(27,10,27,'build','docker-build','success',0,'','/tmp/dxlogs/1787074743149915293.log','2026-08-19 01:39:03.528','2026-08-19 01:39:03.531','2026-08-19 01:39:03.150'),(28,10,28,'push','docker-push','success',0,'','/tmp/dxlogs/1787074743153125757.log','2026-08-19 01:39:04.524','2026-08-19 01:39:04.527','2026-08-19 01:39:03.153'),(29,10,29,'deploy','docker-deploy','failed',-1,'','/tmp/dxlogs/1787074743154893930.log','2026-08-19 01:39:04.715','2026-08-19 01:39:04.719','2026-08-19 01:39:03.155'),(30,11,30,'prepare','shell','success',0,'082fd46875ae2ed877b39e7cc963b475ca18b11b978739c433ae5634ed6ddeff','/tmp/dxlogs/1787074819075172490.log','2026-08-19 01:40:19.097','2026-08-19 01:40:19.470','2026-08-19 01:40:19.075'),(31,11,31,'build','docker-build','success',0,'','/tmp/dxlogs/1787074819077961166.log','2026-08-19 01:40:19.474','2026-08-19 01:40:19.478','2026-08-19 01:40:19.078'),(32,11,32,'push','docker-push','success',0,'','/tmp/dxlogs/1787074819080705620.log','2026-08-19 01:40:19.855','2026-08-19 01:40:19.860','2026-08-19 01:40:19.081'),(33,11,33,'deploy','docker-deploy','failed',-1,'','/tmp/dxlogs/1787074819082809219.log','2026-08-19 01:40:19.956','2026-08-19 01:40:19.963','2026-08-19 01:40:19.083'),(34,12,34,'prepare','shell','success',0,'748d4696fa8c9ce236390145e4d292e4a858b362a84b558c1420b1d132b39ddd','/tmp/dxlogs/1787074876158183912.log','2026-08-19 01:41:16.176','2026-08-19 01:41:16.494','2026-08-19 01:41:16.158'),(35,12,35,'build','docker-build','success',0,'','/tmp/dxlogs/1787074876160036087.log','2026-08-19 01:41:16.499','2026-08-19 01:41:16.503','2026-08-19 01:41:16.160'),(36,12,36,'push','docker-push','success',0,'','/tmp/dxlogs/1787074876162700232.log','2026-08-19 01:41:17.065','2026-08-19 01:41:17.070','2026-08-19 01:41:16.163'),(37,12,37,'deploy','docker-deploy','failed',-1,'','/tmp/dxlogs/1787074876164787919.log','2026-08-19 01:41:17.174','2026-08-19 01:41:17.177','2026-08-19 01:41:16.165'),(38,13,38,'prepare','shell','success',0,'77b1937b4e9d483a2e2f140088e86213bcd3eb8fc1fe286db094a8a39a51110c','/tmp/dxlogs/1787075051896626181.log','2026-08-19 01:44:11.920','2026-08-19 01:44:12.282','2026-08-19 01:44:11.897'),(39,13,39,'build','docker-build','success',0,'','/tmp/dxlogs/1787075051900265866.log','2026-08-19 01:44:12.286','2026-08-19 01:44:12.290','2026-08-19 01:44:11.900'),(40,13,40,'push','docker-push','success',0,'','/tmp/dxlogs/1787075051902560307.log','2026-08-19 01:44:12.882','2026-08-19 01:44:12.886','2026-08-19 01:44:11.903'),(41,13,41,'deploy','docker-deploy','failed',-1,'','/tmp/dxlogs/1787075051907242015.log','2026-08-19 01:44:12.989','2026-08-19 01:44:12.993','2026-08-19 01:44:11.907'),(42,14,42,'cat-dockerfile','shell','success',0,'85448b5dd76b28c8f9a4bb591c31734994ca2ff80d71b0592986d307bdecdd58','/tmp/dxlogs/1787075202458374654.log','2026-08-19 01:46:42.472','2026-08-19 01:46:42.846','2026-08-19 01:46:42.459'),(43,15,43,'prepare-and-cat','shell','success',0,'1a06d79197eaa7f2c3b085fe2aa267ddadc33deedc2c89154b6bf095f27d6d0b','/tmp/dxlogs/1787075223200635186.log','2026-08-19 01:47:03.213','2026-08-19 01:47:03.598','2026-08-19 01:47:03.201'),(44,16,44,'prepare','shell','success',0,'acad2a449ef3c55df3915ca7d1e3bd7318e0882cd726afd987f0056dd2f6fff9','/tmp/dxlogs/1787075283681916842.log','2026-08-19 01:48:03.702','2026-08-19 01:48:04.085','2026-08-19 01:48:03.682'),(45,16,45,'build','docker-build','success',0,'','/tmp/dxlogs/1787075283684042983.log','2026-08-19 01:48:04.089','2026-08-19 01:48:04.093','2026-08-19 01:48:03.684'),(46,16,46,'push','docker-push','success',0,'','/tmp/dxlogs/1787075283687028853.log','2026-08-19 01:48:04.686','2026-08-19 01:48:04.691','2026-08-19 01:48:03.687'),(47,16,47,'deploy','docker-deploy','failed',-1,'','/tmp/dxlogs/1787075283689064122.log','2026-08-19 01:48:04.796','2026-08-19 01:48:04.800','2026-08-19 01:48:03.689'),(48,17,48,'prepare','shell','success',0,'5e481387c11cfa2b593857241d38fc555b0f7c0e97b89ff9488f50372ca1c0d0','/tmp/dxlogs/1787075386393383820.log','2026-08-19 01:49:46.417','2026-08-19 01:49:46.809','2026-08-19 01:49:46.394'),(49,17,49,'build','docker-build','success',0,'','/tmp/dxlogs/1787075386396541654.log','2026-08-19 01:49:46.813','2026-08-19 01:49:46.818','2026-08-19 01:49:46.397'),(50,17,50,'push','docker-push','success',0,'','/tmp/dxlogs/1787075386399308279.log','2026-08-19 01:49:47.529','2026-08-19 01:49:47.533','2026-08-19 01:49:46.399'),(51,17,51,'deploy','docker-deploy','success',0,'','/tmp/dxlogs/1787075386401726129.log','2026-08-19 01:49:47.636','2026-08-19 01:49:47.639','2026-08-19 01:49:46.402'),(52,18,52,'prepare','shell','success',0,'7078b135a7b0d180bb443f0f55483c368d5b677e659bd341b5fa72accb4b8c2d','/tmp/dxlogs/1787075410146363827.log','2026-08-19 01:50:10.168','2026-08-19 01:50:10.571','2026-08-19 01:50:10.147'),(53,18,53,'build','docker-build','success',0,'','/tmp/dxlogs/1787075410149467416.log','2026-08-19 01:50:10.580','2026-08-19 01:50:10.585','2026-08-19 01:50:10.150'),(54,18,54,'push','docker-push','success',0,'','/tmp/dxlogs/1787075410151699626.log','2026-08-19 01:50:11.036','2026-08-19 01:50:11.041','2026-08-19 01:50:10.152'),(55,18,55,'deploy','docker-deploy','success',0,'','/tmp/dxlogs/1787075410155034595.log','2026-08-19 01:50:11.128','2026-08-19 01:50:11.134','2026-08-19 01:50:10.155'),(56,19,56,'prepare','shell','success',0,'1aef63178c931df87a7ab7d7037ae395218632cc01312610b28cd702a6bfcd12','/tmp/dxlogs/1787076501734777678.log','2026-08-19 02:08:21.763','2026-08-19 02:08:22.111','2026-08-19 02:08:21.735'),(57,19,57,'build','docker-build','success',0,'','/tmp/dxlogs/1787076501739501835.log','2026-08-19 02:08:22.115','2026-08-19 02:08:22.119','2026-08-19 02:08:21.740'),(58,19,58,'push','docker-push','success',0,'','/tmp/dxlogs/1787076501742249334.log','2026-08-19 02:08:22.495','2026-08-19 02:08:22.500','2026-08-19 02:08:21.742'),(59,19,59,'deploy','docker-deploy','success',0,'','/tmp/dxlogs/1787076501746546804.log','2026-08-19 02:08:22.589','2026-08-19 02:08:22.593','2026-08-19 02:08:21.747');
/*!40000 ALTER TABLE `pipeline_job_runs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `pipeline_runs`
--

DROP TABLE IF EXISTS `pipeline_runs`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pipeline_runs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint unsigned NOT NULL,
  `run_no` int NOT NULL DEFAULT '1',
  `trigger_type` varchar(16) NOT NULL DEFAULT 'manual' COMMENT 'manual/webhook',
  `ref` varchar(64) NOT NULL DEFAULT '',
  `commit_sha` varchar(64) NOT NULL DEFAULT '',
  `status` varchar(16) NOT NULL DEFAULT 'pending' COMMENT 'pending/running/success/failed/canceled',
  `started_at` datetime(3) DEFAULT NULL,
  `finished_at` datetime(3) DEFAULT NULL,
  `duration_ms` bigint NOT NULL DEFAULT '0',
  `triggered_by` bigint unsigned DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_pipeline_runs_pipe` (`pipeline_id`,`status`),
  KEY `idx_pipeline_runs_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=20 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Pipeline 执行记录';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `pipeline_runs`
--

LOCK TABLES `pipeline_runs` WRITE;
/*!40000 ALTER TABLE `pipeline_runs` DISABLE KEYS */;
INSERT INTO `pipeline_runs` VALUES (1,1,1,'manual','','','success','2026-08-19 01:24:32.872','2026-08-19 01:24:33.911',1038,1,'2026-08-19 01:24:32.858','2026-08-19 01:24:33.911'),(2,2,1,'manual','','','failed','2026-08-19 01:24:53.066','2026-08-19 01:24:53.464',398,1,'2026-08-19 01:24:53.054','2026-08-19 01:24:53.464'),(3,3,1,'manual','','','failed','2026-08-19 01:24:58.124','2026-08-19 01:25:06.279',8156,1,'2026-08-19 01:24:58.116','2026-08-19 01:25:06.279'),(4,3,2,'manual','','','failed','2026-08-19 01:26:30.744','2026-08-19 01:26:38.906',8163,1,'2026-08-19 01:26:30.732','2026-08-19 01:26:38.906'),(5,2,2,'manual','','','failed','2026-08-19 01:26:43.917','2026-08-19 01:26:44.253',335,1,'2026-08-19 01:26:43.910','2026-08-19 01:26:44.253'),(6,3,3,'manual','','','canceled','2026-08-19 01:28:00.440','2026-08-19 01:28:08.618',8179,1,'2026-08-19 01:28:00.430','2026-08-19 01:28:08.619'),(7,4,1,'webhook','main','abc123def456789','failed','2026-08-19 01:34:28.678','2026-08-19 01:34:29.411',733,1,'2026-08-19 01:34:28.665','2026-08-19 01:34:29.412'),(8,4,2,'manual','','','failed','2026-08-19 01:36:15.809','2026-08-19 01:36:16.516',707,1,'2026-08-19 01:36:15.794','2026-08-19 01:36:16.516'),(9,4,3,'manual','','','failed','2026-08-19 01:37:36.105','2026-08-19 01:37:36.889',783,1,'2026-08-19 01:37:36.091','2026-08-19 01:37:36.889'),(10,4,4,'manual','','','failed','2026-08-19 01:39:03.158','2026-08-19 01:39:04.918',1760,1,'2026-08-19 01:39:03.145','2026-08-19 01:39:04.919'),(11,4,5,'manual','','','failed','2026-08-19 01:40:19.088','2026-08-19 01:40:20.139',1050,1,'2026-08-19 01:40:19.073','2026-08-19 01:40:20.139'),(12,4,6,'manual','','','failed','2026-08-19 01:41:16.168','2026-08-19 01:42:17.681',61517,1,'2026-08-19 01:41:16.156','2026-08-19 01:42:17.681'),(13,4,7,'manual','','','failed','2026-08-19 01:44:11.912','2026-08-19 01:45:13.504',61595,1,'2026-08-19 01:44:11.894','2026-08-19 01:45:13.505'),(14,5,1,'manual','','','success','2026-08-19 01:46:42.462','2026-08-19 01:46:42.850',388,1,'2026-08-19 01:46:42.456','2026-08-19 01:46:42.851'),(15,6,1,'manual','','','success','2026-08-19 01:47:03.204','2026-08-19 01:47:03.602',397,1,'2026-08-19 01:47:03.198','2026-08-19 01:47:03.602'),(16,4,8,'manual','','','failed','2026-08-19 01:48:03.693','2026-08-19 01:49:05.371',61682,1,'2026-08-19 01:48:03.679','2026-08-19 01:49:05.372'),(17,4,9,'manual','','','success','2026-08-19 01:49:46.406','2026-08-19 01:49:47.889',1482,1,'2026-08-19 01:49:46.390','2026-08-19 01:49:47.890'),(18,4,10,'webhook','release/1.0','bbbb2222','success','2026-08-19 01:50:10.159','2026-08-19 01:50:11.919',1760,1,'2026-08-19 01:50:10.144','2026-08-19 01:50:11.920'),(19,4,11,'manual','','','success','2026-08-19 02:08:21.754','2026-08-19 02:08:23.299',1545,1,'2026-08-19 02:08:21.731','2026-08-19 02:08:23.300');
/*!40000 ALTER TABLE `pipeline_runs` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `pipeline_steps`
--

DROP TABLE IF EXISTS `pipeline_steps`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pipeline_steps` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `pipeline_id` bigint unsigned NOT NULL,
  `name` varchar(64) NOT NULL,
  `type` varchar(32) NOT NULL,
  `seq` int NOT NULL DEFAULT '0',
  `config_json` text,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pipeline_steps` (`pipeline_id`,`seq`),
  KEY `idx_pipeline_steps_pipe` (`pipeline_id`)
) ENGINE=InnoDB AUTO_INCREMENT=60 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Pipeline 步骤（解析快照）';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `pipeline_steps`
--

LOCK TABLES `pipeline_steps` WRITE;
/*!40000 ALTER TABLE `pipeline_steps` DISABLE KEYS */;
INSERT INTO `pipeline_steps` VALUES (1,1,'greet','shell',1,'{\"Name\":\"greet\",\"Type\":\"shell\",\"Script\":\"echo hello-pipeline \\u0026\\u0026 pwd \\u0026\\u0026 ls -la\",\"URL\":\"\",\"Branch\":\"\",\"AllowFailure\":false,\"Timeout\":\"\",\"Dockerfile\":\"\",\"Tags\":null,\"Application\":\"\",\"Environment\":\"\"}','2026-08-19 01:24:32.850'),(2,1,'tolerate-fail','shell',2,'{\"Name\":\"tolerate-fail\",\"Type\":\"shell\",\"Script\":\"echo about-to-fail \\u0026\\u0026 exit 3\",\"URL\":\"\",\"Branch\":\"\",\"AllowFailure\":true,\"Timeout\":\"\",\"Dockerfile\":\"\",\"Tags\":null,\"Application\":\"\",\"Environment\":\"\"}','2026-08-19 01:24:32.850'),(3,1,'final','shell',3,'{\"Name\":\"final\",\"Type\":\"shell\",\"Script\":\"echo all-done\",\"URL\":\"\",\"Branch\":\"\",\"AllowFailure\":false,\"Timeout\":\"\",\"Dockerfile\":\"\",\"Tags\":null,\"Application\":\"\",\"Environment\":\"\"}','2026-08-19 01:24:32.851'),(10,2,'boom','shell',1,'{\"Name\":\"boom\",\"Type\":\"shell\",\"Script\":\"echo before-boom \\u0026\\u0026 exit 7\",\"URL\":\"\",\"Branch\":\"\",\"AllowFailure\":false,\"Timeout\":\"\",\"Dockerfile\":\"\",\"Tags\":null,\"Application\":\"\",\"Environment\":\"\"}','2026-08-19 01:26:43.905'),(11,2,'never','shell',2,'{\"Name\":\"never\",\"Type\":\"shell\",\"Script\":\"echo should-not-run\",\"URL\":\"\",\"Branch\":\"\",\"AllowFailure\":false,\"Timeout\":\"\",\"Dockerfile\":\"\",\"Tags\":null,\"Application\":\"\",\"Environment\":\"\"}','2026-08-19 01:26:43.905'),(12,3,'long-sleep','shell',1,'{\"Name\":\"long-sleep\",\"Type\":\"shell\",\"Script\":\"echo start-sleep \\u0026\\u0026 sleep 60 \\u0026\\u0026 echo woke-up\",\"URL\":\"\",\"Branch\":\"\",\"AllowFailure\":false,\"Timeout\":\"\",\"Dockerfile\":\"\",\"Tags\":null,\"Application\":\"\",\"Environment\":\"\"}','2026-08-19 01:28:00.425'),(13,3,'after','shell',2,'{\"Name\":\"after\",\"Type\":\"shell\",\"Script\":\"echo after-sleep\",\"URL\":\"\",\"Branch\":\"\",\"AllowFailure\":false,\"Timeout\":\"\",\"Dockerfile\":\"\",\"Tags\":null,\"Application\":\"\",\"Environment\":\"\"}','2026-08-19 01:28:00.425'),(42,5,'cat-dockerfile','shell',1,'{\"Name\":\"cat-dockerfile\",\"Type\":\"shell\",\"Script\":\"echo \'=== Dockerfile ???? ===\'\\ncat Dockerfile\\necho \'=== ???? ===\'\",\"URL\":\"\",\"Branch\":\"\",\"AllowFailure\":false,\"Timeout\":\"\",\"Dockerfile\":\"\",\"Tags\":null,\"Image\":\"\",\"Application\":\"\",\"Environment\":\"\"}','2026-08-19 01:46:42.451'),(43,6,'prepare-and-cat','shell',1,'{\"Name\":\"prepare-and-cat\",\"Type\":\"shell\",\"Script\":\"cat \\u003e Dockerfile \\u003c\\u003c\'EOF\'\\nFROM alpine:3.20\\nCMD [\\\"sh\\\",\\\"-c\\\",\\\"while true; do { printf \'HTTP/1.1 200 OK\\\\015\\\\012Content-Length: 12\\\\015\\\\012\\\\015\\\\012\\u003ch1\\u003epipe-ok\\u003c/h1\\u003e\'; } | nc -l -p 80 -q 1; done\\\"]\\nEOF\\necho ---BEGIN---\\ncat -A Dockerfile\\necho ---END---\",\"URL\":\"\",\"Branch\":\"\",\"AllowFailure\":false,\"Timeout\":\"\",\"Dockerfile\":\"\",\"Tags\":null,\"Image\":\"\",\"Application\":\"\",\"Environment\":\"\"}','2026-08-19 01:47:03.193'),(56,4,'prepare','shell',1,'{\"Name\":\"prepare\",\"Type\":\"shell\",\"Script\":\"cat \\u003e resp.sh \\u003c\\u003c\'EOF\'\\n#!/bin/sh\\nprintf \'HTTP/1.1 200 OK\\\\r\\\\nContent-Length: 12\\\\r\\\\n\\\\r\\\\n\\u003ch1\\u003epipe-ok\\u003c/h1\\u003e\'\\nEOF\\nchmod +x resp.sh\\ncat \\u003e Dockerfile \\u003c\\u003c\'EOF\'\\nFROM alpine:3.20\\nCOPY resp.sh /resp.sh\\nCMD [\\\"nc\\\",\\\"-lk\\\",\\\"-p\\\",\\\"80\\\",\\\"-e\\\",\\\"/resp.sh\\\"]\\nEOF\\n\",\"URL\":\"\",\"Branch\":\"\",\"AllowFailure\":false,\"Timeout\":\"\",\"Dockerfile\":\"\",\"Tags\":null,\"Image\":\"\",\"Application\":\"\",\"Environment\":\"\"}','2026-08-19 02:08:21.726'),(57,4,'build','docker-build',2,'{\"Name\":\"build\",\"Type\":\"docker-build\",\"Script\":\"\",\"URL\":\"\",\"Branch\":\"\",\"AllowFailure\":false,\"Timeout\":\"\",\"Dockerfile\":\"Dockerfile\",\"Tags\":[\"host.docker.internal:15000/default/pipetest:v5\"],\"Image\":\"\",\"Application\":\"\",\"Environment\":\"\"}','2026-08-19 02:08:21.726'),(58,4,'push','docker-push',3,'{\"Name\":\"push\",\"Type\":\"docker-push\",\"Script\":\"\",\"URL\":\"\",\"Branch\":\"\",\"AllowFailure\":false,\"Timeout\":\"\",\"Dockerfile\":\"\",\"Tags\":[\"host.docker.internal:15000/default/pipetest:v5\"],\"Image\":\"\",\"Application\":\"\",\"Environment\":\"\"}','2026-08-19 02:08:21.727'),(59,4,'deploy','docker-deploy',4,'{\"Name\":\"deploy\",\"Type\":\"docker-deploy\",\"Script\":\"\",\"URL\":\"\",\"Branch\":\"\",\"AllowFailure\":false,\"Timeout\":\"\",\"Dockerfile\":\"\",\"Tags\":null,\"Image\":\"host.docker.internal:15000/default/pipetest:v5\",\"Application\":\"pipe-app\",\"Environment\":\"\"}','2026-08-19 02:08:21.727');
/*!40000 ALTER TABLE `pipeline_steps` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `pipelines`
--

DROP TABLE IF EXISTS `pipelines`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pipelines` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned DEFAULT NULL,
  `project_id` bigint unsigned DEFAULT NULL,
  `owner_id` bigint unsigned NOT NULL DEFAULT '0',
  `name` varchar(64) NOT NULL,
  `description` varchar(255) NOT NULL DEFAULT '',
  `definition` mediumtext COMMENT 'Pipeline YAML 定义',
  `status` tinyint NOT NULL DEFAULT '1',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pipelines_name` (`name`),
  KEY `idx_pipelines_project` (`project_id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Pipeline 定义';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `pipelines`
--

LOCK TABLES `pipelines` WRITE;
/*!40000 ALTER TABLE `pipelines` DISABLE KEYS */;
INSERT INTO `pipelines` VALUES (1,NULL,NULL,1,'hello-pipe','','name: hello-pipe\nsteps:\n  - name: greet\n    type: shell\n    script: echo hello-pipeline && pwd && ls -la\n  - name: tolerate-fail\n    type: shell\n    script: echo about-to-fail && exit 3\n    allow_failure: true\n  - name: final\n    type: shell\n    script: echo all-done',1,'2026-08-19 01:24:32.832','2026-08-19 01:50:21.195','2026-08-19 01:50:21.196'),(2,NULL,NULL,1,'fail-pipe','','name: fail-pipe\nsteps:\n  - name: boom\n    type: shell\n    script: echo before-boom && exit 7\n  - name: never\n    type: shell\n    script: echo should-not-run',1,'2026-08-19 01:24:53.020','2026-08-19 01:50:21.186','2026-08-19 01:50:21.186'),(3,NULL,NULL,1,'cancel-pipe','','name: cancel-pipe\nsteps:\n  - name: long-sleep\n    type: shell\n    script: echo start-sleep && sleep 60 && echo woke-up\n  - name: after\n    type: shell\n    script: echo after-sleep',1,'2026-08-19 01:24:58.100','2026-08-19 01:50:21.178','2026-08-19 01:50:21.178'),(4,NULL,NULL,1,'ci-cd','','name: ci-cd\nsteps:\n  - name: prepare\n    type: shell\n    script: |\n      cat > resp.sh <<\'EOF\'\n      #!/bin/sh\n      printf \'HTTP/1.1 200 OK\\r\\nContent-Length: 12\\r\\n\\r\\n<h1>pipe-ok</h1>\'\n      EOF\n      chmod +x resp.sh\n      cat > Dockerfile <<\'EOF\'\n      FROM alpine:3.20\n      COPY resp.sh /resp.sh\n      CMD [\"nc\",\"-lk\",\"-p\",\"80\",\"-e\",\"/resp.sh\"]\n      EOF\n  - name: build\n    type: docker-build\n    dockerfile: Dockerfile\n    tags:\n      - host.docker.internal:15000/default/pipetest:v5\n  - name: push\n    type: docker-push\n    tags:\n      - host.docker.internal:15000/default/pipetest:v5\n  - name: deploy\n    type: docker-deploy\n    application: pipe-app\n    image: host.docker.internal:15000/default/pipetest:v5',1,'2026-08-19 01:34:18.006','2026-08-19 01:49:46.371',NULL),(5,NULL,NULL,1,'debug-cat','','name: debug-cat\nsteps:\n  - name: cat-dockerfile\n    type: shell\n    script: |\n      echo \'=== Dockerfile ???? ===\'\n      cat Dockerfile\n      echo \'=== ???? ===\'',1,'2026-08-19 01:46:42.418','2026-08-19 01:50:21.159','2026-08-19 01:50:21.160'),(6,NULL,NULL,1,'debug-cat2','','name: debug-cat2\nsteps:\n  - name: prepare-and-cat\n    type: shell\n    script: |\n      cat > Dockerfile <<\'EOF\'\n      FROM alpine:3.20\n      CMD [\"sh\",\"-c\",\"while true; do { printf \'HTTP/1.1 200 OK\\015\\012Content-Length: 12\\015\\012\\015\\012<h1>pipe-ok</h1>\'; } | nc -l -p 80 -q 1; done\"]\n      EOF\n      echo ---BEGIN---\n      cat -A Dockerfile\n      echo ---END---',1,'2026-08-19 01:47:03.162','2026-08-19 01:50:21.150','2026-08-19 01:50:21.150');
/*!40000 ALTER TABLE `pipelines` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `project_environments`
--

DROP TABLE IF EXISTS `project_environments`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `project_environments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `project_id` bigint unsigned NOT NULL,
  `name` varchar(16) NOT NULL COMMENT 'development/testing/staging/production',
  `seq` int NOT NULL DEFAULT '0',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_project_envs` (`project_id`,`name`)
) ENGINE=InnoDB AUTO_INCREMENT=21 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='项目环境';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `project_environments`
--

LOCK TABLES `project_environments` WRITE;
/*!40000 ALTER TABLE `project_environments` DISABLE KEYS */;
INSERT INTO `project_environments` VALUES (1,1,'development',1,'2026-08-19 01:06:58.998'),(2,1,'testing',2,'2026-08-19 01:06:59.003'),(3,1,'staging',3,'2026-08-19 01:06:59.005'),(4,1,'production',4,'2026-08-19 01:06:59.007'),(5,2,'development',1,'2026-08-19 02:29:51.155'),(6,2,'testing',2,'2026-08-19 02:29:51.163'),(7,2,'staging',3,'2026-08-19 02:29:51.170'),(8,2,'production',4,'2026-08-19 02:29:51.175'),(9,3,'development',1,'2026-08-19 02:29:51.207'),(10,3,'testing',2,'2026-08-19 02:29:51.213'),(11,3,'staging',3,'2026-08-19 02:29:51.220'),(12,3,'production',4,'2026-08-19 02:29:51.226'),(13,4,'development',1,'2026-08-19 02:33:11.151'),(14,4,'testing',2,'2026-08-19 02:33:11.155'),(15,4,'staging',3,'2026-08-19 02:33:11.159'),(16,4,'production',4,'2026-08-19 02:33:11.163'),(17,5,'development',1,'2026-08-19 02:33:11.186'),(18,5,'testing',2,'2026-08-19 02:33:11.190'),(19,5,'staging',3,'2026-08-19 02:33:11.194'),(20,5,'production',4,'2026-08-19 02:33:11.198');
/*!40000 ALTER TABLE `project_environments` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `projects`
--

DROP TABLE IF EXISTS `projects`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `projects` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned NOT NULL,
  `name` varchar(128) NOT NULL,
  `code` varchar(64) NOT NULL,
  `description` varchar(255) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT '1',
  `created_by` bigint unsigned DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_projects_org_name` (`org_id`,`name`),
  KEY `idx_projects_org` (`org_id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='项目';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `projects`
--

LOCK TABLES `projects` WRITE;
/*!40000 ALTER TABLE `projects` DISABLE KEYS */;
INSERT INTO `projects` VALUES (1,0,'shop','shop','',1,1,'2026-08-19 01:06:58.994','2026-08-19 01:06:58.994',NULL),(2,1,'pa-web-0819022950','pa-web','',1,1,'2026-08-19 02:29:51.146','2026-08-19 02:29:51.146',NULL),(3,2,'pb-web-0819022950','pb-web','',1,1,'2026-08-19 02:29:51.199','2026-08-19 02:29:51.199',NULL),(4,3,'pa-web-0819023310','pa-web','',1,1,'2026-08-19 02:33:11.147','2026-08-19 02:33:11.147',NULL),(5,4,'pb-web-0819023310','pb-web','',1,1,'2026-08-19 02:33:11.175','2026-08-19 02:33:11.175',NULL);
/*!40000 ALTER TABLE `projects` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `registries`
--

DROP TABLE IF EXISTS `registries`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `registries` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned DEFAULT NULL,
  `name` varchar(64) NOT NULL,
  `url` varchar(255) NOT NULL,
  `username` varchar(128) NOT NULL DEFAULT '',
  `password_enc` varchar(255) NOT NULL DEFAULT '',
  `type` varchar(16) NOT NULL DEFAULT 'self' COMMENT 'self/docker-hub/other',
  `status` tinyint NOT NULL DEFAULT '1',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_registries_org_name` (`org_id`,`name`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='镜像仓库源';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `registries`
--

LOCK TABLES `registries` WRITE;
/*!40000 ALTER TABLE `registries` DISABLE KEYS */;
INSERT INTO `registries` VALUES (1,NULL,'内置 Registry','registry:5000','','','self',1,'2026-08-19 00:47:13.857','2026-08-19 00:47:13.857',NULL);
/*!40000 ALTER TABLE `registries` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `registry_repositories`
--

DROP TABLE IF EXISTS `registry_repositories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `registry_repositories` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `registry_id` bigint unsigned NOT NULL,
  `org_id` bigint unsigned DEFAULT NULL,
  `project_id` bigint unsigned DEFAULT NULL,
  `namespace` varchar(128) NOT NULL DEFAULT '',
  `name` varchar(255) NOT NULL,
  `visibility` varchar(16) NOT NULL DEFAULT 'private' COMMENT 'private/public',
  `pull_count` bigint NOT NULL DEFAULT '0',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_registry_repos` (`registry_id`,`namespace`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='仓库（namespace/name）';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `registry_repositories`
--

LOCK TABLES `registry_repositories` WRITE;
/*!40000 ALTER TABLE `registry_repositories` DISABLE KEYS */;
/*!40000 ALTER TABLE `registry_repositories` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `resource_quotas`
--

DROP TABLE IF EXISTS `resource_quotas`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `resource_quotas` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned NOT NULL,
  `project_id` bigint unsigned DEFAULT NULL,
  `resource_type` varchar(32) NOT NULL COMMENT 'ecs_count/cpu/memory/storage/network/pipeline',
  `limit_value` bigint NOT NULL DEFAULT '0',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_resource_quotas` (`org_id`,`project_id`,`resource_type`),
  KEY `idx_resource_quotas_org` (`org_id`)
) ENGINE=InnoDB AUTO_INCREMENT=25 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='资源配额';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `resource_quotas`
--

LOCK TABLES `resource_quotas` WRITE;
/*!40000 ALTER TABLE `resource_quotas` DISABLE KEYS */;
INSERT INTO `resource_quotas` VALUES (1,1,NULL,'memory',16384,'2026-08-19 02:29:50.746','2026-08-19 02:29:50.746'),(2,1,NULL,'storage',100,'2026-08-19 02:29:50.753','2026-08-19 02:29:50.753'),(3,1,NULL,'network',10,'2026-08-19 02:29:50.758','2026-08-19 02:29:50.758'),(4,1,NULL,'pipeline',10,'2026-08-19 02:29:50.765','2026-08-19 02:29:50.765'),(5,1,NULL,'ecs_count',10,'2026-08-19 02:29:50.771','2026-08-19 02:29:54.942'),(6,1,NULL,'cpu',8,'2026-08-19 02:29:50.776','2026-08-19 02:29:50.776'),(7,2,NULL,'network',10,'2026-08-19 02:29:50.807','2026-08-19 02:29:50.807'),(8,2,NULL,'pipeline',10,'2026-08-19 02:29:50.813','2026-08-19 02:29:50.813'),(9,2,NULL,'ecs_count',5,'2026-08-19 02:29:50.818','2026-08-19 02:29:50.818'),(10,2,NULL,'cpu',8,'2026-08-19 02:29:50.824','2026-08-19 02:29:50.824'),(11,2,NULL,'memory',16384,'2026-08-19 02:29:50.831','2026-08-19 02:29:50.831'),(12,2,NULL,'storage',100,'2026-08-19 02:29:50.836','2026-08-19 02:29:50.836'),(13,3,NULL,'memory',16384,'2026-08-19 02:33:10.875','2026-08-19 02:33:10.875'),(14,3,NULL,'storage',100,'2026-08-19 02:33:10.879','2026-08-19 02:33:10.879'),(15,3,NULL,'network',10,'2026-08-19 02:33:10.883','2026-08-19 02:33:10.883'),(16,3,NULL,'pipeline',10,'2026-08-19 02:33:10.887','2026-08-19 02:33:10.887'),(17,3,NULL,'ecs_count',10,'2026-08-19 02:33:10.890','2026-08-19 02:33:14.644'),(18,3,NULL,'cpu',8,'2026-08-19 02:33:10.893','2026-08-19 02:33:10.893'),(19,4,NULL,'memory',16384,'2026-08-19 02:33:10.913','2026-08-19 02:33:10.913'),(20,4,NULL,'storage',100,'2026-08-19 02:33:10.916','2026-08-19 02:33:10.916'),(21,4,NULL,'network',10,'2026-08-19 02:33:10.920','2026-08-19 02:33:10.920'),(22,4,NULL,'pipeline',10,'2026-08-19 02:33:10.924','2026-08-19 02:33:10.924'),(23,4,NULL,'ecs_count',5,'2026-08-19 02:33:10.927','2026-08-19 02:33:10.927'),(24,4,NULL,'cpu',8,'2026-08-19 02:33:10.931','2026-08-19 02:33:10.931');
/*!40000 ALTER TABLE `resource_quotas` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `resource_usage`
--

DROP TABLE IF EXISTS `resource_usage`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `resource_usage` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned NOT NULL,
  `project_id` bigint unsigned DEFAULT NULL,
  `resource_type` varchar(32) NOT NULL COMMENT 'cpu_hour/mem_gb_hour/disk_gb_hour',
  `used_value` decimal(18,4) NOT NULL DEFAULT '0.0000',
  `period` datetime(3) NOT NULL COMMENT '所属小时（整点）',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_resource_usage_org` (`org_id`,`period`),
  KEY `idx_resource_usage_period` (`period`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='资源用量（虚拟计费）';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `resource_usage`
--

LOCK TABLES `resource_usage` WRITE;
/*!40000 ALTER TABLE `resource_usage` DISABLE KEYS */;
INSERT INTO `resource_usage` VALUES (1,1,NULL,'cpu_hour',1.0000,'2026-08-19 02:00:00.000','2026-08-19 02:29:54.831'),(2,1,NULL,'mem_gb_hour',0.5000,'2026-08-19 02:00:00.000','2026-08-19 02:29:54.838'),(3,1,NULL,'disk_gb_hour',10.0000,'2026-08-19 02:00:00.000','2026-08-19 02:29:54.845'),(4,3,NULL,'cpu_hour',1.0000,'2026-08-19 02:00:00.000','2026-08-19 02:33:14.563'),(5,3,NULL,'mem_gb_hour',0.5000,'2026-08-19 02:00:00.000','2026-08-19 02:33:14.566'),(6,3,NULL,'disk_gb_hour',10.0000,'2026-08-19 02:00:00.000','2026-08-19 02:33:14.569'),(7,1,NULL,'cpu_hour',2.0000,'2026-08-19 02:00:00.000','2026-08-19 02:33:14.575'),(8,1,NULL,'mem_gb_hour',1.0000,'2026-08-19 02:00:00.000','2026-08-19 02:33:14.578'),(9,1,NULL,'disk_gb_hour',20.0000,'2026-08-19 02:00:00.000','2026-08-19 02:33:14.581');
/*!40000 ALTER TABLE `resource_usage` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `role_permissions`
--

DROP TABLE IF EXISTS `role_permissions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `role_permissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `role_id` bigint unsigned NOT NULL,
  `permission_id` bigint unsigned NOT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_permissions` (`role_id`,`permission_id`)
) ENGINE=InnoDB AUTO_INCREMENT=272 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='角色-权限';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `role_permissions`
--

LOCK TABLES `role_permissions` WRITE;
/*!40000 ALTER TABLE `role_permissions` DISABLE KEYS */;
INSERT INTO `role_permissions` VALUES (1,3,1,'2026-08-18 23:46:05.205'),(2,3,2,'2026-08-18 23:46:05.208'),(3,3,3,'2026-08-18 23:46:05.212'),(4,3,4,'2026-08-18 23:46:05.214'),(5,3,5,'2026-08-18 23:46:05.218'),(6,3,6,'2026-08-18 23:46:05.220'),(7,3,7,'2026-08-18 23:46:05.223'),(8,3,8,'2026-08-18 23:46:05.226'),(9,3,9,'2026-08-18 23:46:05.228'),(10,3,10,'2026-08-18 23:46:05.232'),(11,3,11,'2026-08-18 23:46:05.260'),(12,3,12,'2026-08-18 23:46:05.266'),(13,3,13,'2026-08-18 23:46:05.270'),(14,3,14,'2026-08-18 23:46:05.274'),(15,3,15,'2026-08-18 23:46:05.278'),(16,3,16,'2026-08-18 23:46:05.282'),(17,3,17,'2026-08-18 23:46:05.286'),(18,3,18,'2026-08-18 23:46:05.289'),(19,3,19,'2026-08-18 23:46:05.293'),(20,3,20,'2026-08-18 23:46:05.297'),(21,3,21,'2026-08-18 23:46:05.302'),(22,3,22,'2026-08-18 23:46:05.305'),(23,3,23,'2026-08-18 23:46:05.308'),(24,3,24,'2026-08-18 23:46:05.312'),(25,3,25,'2026-08-18 23:46:05.315'),(26,3,26,'2026-08-18 23:46:05.340'),(27,3,27,'2026-08-18 23:46:05.343'),(28,3,28,'2026-08-18 23:46:05.346'),(29,3,29,'2026-08-18 23:46:05.348'),(30,3,30,'2026-08-18 23:46:05.352'),(31,3,31,'2026-08-18 23:46:05.355'),(32,3,32,'2026-08-18 23:46:05.358'),(33,3,33,'2026-08-18 23:46:05.361'),(34,3,34,'2026-08-18 23:46:05.365'),(35,3,35,'2026-08-18 23:46:05.368'),(36,3,36,'2026-08-18 23:46:05.370'),(37,3,37,'2026-08-18 23:46:05.373'),(38,3,38,'2026-08-18 23:46:05.376'),(39,3,39,'2026-08-18 23:46:05.379'),(40,3,40,'2026-08-18 23:46:05.382'),(41,3,41,'2026-08-18 23:46:05.384'),(42,3,42,'2026-08-18 23:46:05.387'),(43,3,43,'2026-08-18 23:46:05.390'),(44,3,44,'2026-08-18 23:46:05.393'),(45,3,45,'2026-08-18 23:46:05.396'),(46,3,46,'2026-08-18 23:46:05.399'),(47,3,47,'2026-08-18 23:46:05.402'),(48,3,48,'2026-08-18 23:46:05.405'),(49,3,49,'2026-08-18 23:46:05.408'),(50,3,50,'2026-08-18 23:46:05.411'),(51,3,51,'2026-08-18 23:46:05.413'),(52,3,52,'2026-08-18 23:46:05.417'),(53,3,53,'2026-08-18 23:46:05.419'),(54,4,1,'2026-08-18 23:46:05.422'),(55,4,2,'2026-08-18 23:46:05.426'),(56,4,6,'2026-08-18 23:46:05.429'),(57,4,7,'2026-08-18 23:46:05.432'),(58,4,8,'2026-08-18 23:46:05.435'),(59,4,9,'2026-08-18 23:46:05.437'),(60,4,10,'2026-08-18 23:46:05.440'),(61,4,11,'2026-08-18 23:46:05.443'),(62,4,12,'2026-08-18 23:46:05.466'),(63,4,39,'2026-08-18 23:46:05.470'),(64,4,43,'2026-08-18 23:46:05.473'),(65,4,44,'2026-08-18 23:46:05.476'),(66,4,33,'2026-08-18 23:46:05.480'),(67,4,14,'2026-08-18 23:46:05.482'),(68,4,20,'2026-08-18 23:46:05.485'),(69,4,24,'2026-08-18 23:46:05.489'),(70,4,28,'2026-08-18 23:46:05.492'),(71,4,45,'2026-08-18 23:46:05.495'),(72,4,50,'2026-08-18 23:46:05.498'),(73,5,1,'2026-08-18 23:46:05.501'),(74,5,2,'2026-08-18 23:46:05.504'),(75,5,3,'2026-08-18 23:46:05.507'),(76,5,4,'2026-08-18 23:46:05.510'),(77,5,5,'2026-08-18 23:46:05.536'),(78,5,6,'2026-08-18 23:46:05.541'),(79,5,7,'2026-08-18 23:46:05.544'),(80,5,8,'2026-08-18 23:46:05.548'),(81,5,9,'2026-08-18 23:46:05.552'),(82,5,10,'2026-08-18 23:46:05.556'),(83,5,11,'2026-08-18 23:46:05.559'),(84,5,12,'2026-08-18 23:46:05.562'),(85,5,13,'2026-08-18 23:46:05.565'),(86,5,33,'2026-08-18 23:46:05.590'),(87,5,34,'2026-08-18 23:46:05.595'),(88,5,35,'2026-08-18 23:46:05.599'),(89,5,36,'2026-08-18 23:46:05.604'),(90,5,37,'2026-08-18 23:46:05.613'),(91,5,38,'2026-08-18 23:46:05.617'),(92,5,39,'2026-08-18 23:46:05.642'),(93,5,40,'2026-08-18 23:46:05.645'),(94,5,41,'2026-08-18 23:46:05.649'),(95,5,42,'2026-08-18 23:46:05.651'),(96,5,43,'2026-08-18 23:46:05.654'),(97,5,44,'2026-08-18 23:46:05.656'),(98,5,14,'2026-08-18 23:46:05.659'),(99,5,15,'2026-08-18 23:46:05.662'),(100,5,20,'2026-08-18 23:46:05.666'),(101,5,24,'2026-08-18 23:46:05.669'),(102,5,28,'2026-08-18 23:46:05.672'),(103,5,45,'2026-08-18 23:46:05.675'),(104,5,46,'2026-08-18 23:46:05.677'),(105,5,50,'2026-08-18 23:46:05.680'),(106,6,1,'2026-08-18 23:46:05.684'),(107,6,2,'2026-08-18 23:46:05.686'),(108,6,11,'2026-08-18 23:46:05.712'),(109,6,12,'2026-08-18 23:46:05.715'),(110,6,14,'2026-08-18 23:46:05.718'),(111,6,20,'2026-08-18 23:46:05.720'),(112,6,24,'2026-08-18 23:46:05.768'),(113,6,28,'2026-08-18 23:46:05.777'),(114,6,33,'2026-08-18 23:46:05.793'),(115,6,39,'2026-08-18 23:46:05.798'),(116,6,45,'2026-08-18 23:46:05.801'),(117,6,50,'2026-08-18 23:46:05.803'),(118,6,64,'2026-08-18 23:46:05.806'),(119,6,66,'2026-08-18 23:46:05.809'),(120,6,67,'2026-08-18 23:46:05.812'),(121,1,1,'2026-08-18 23:46:05.814'),(122,1,2,'2026-08-18 23:46:05.816'),(123,1,3,'2026-08-18 23:46:05.819'),(124,1,4,'2026-08-18 23:46:05.822'),(125,1,5,'2026-08-18 23:46:05.825'),(126,1,6,'2026-08-18 23:46:05.829'),(127,1,7,'2026-08-18 23:46:05.832'),(128,1,8,'2026-08-18 23:46:05.836'),(129,1,9,'2026-08-18 23:46:05.839'),(130,1,10,'2026-08-18 23:46:05.843'),(131,1,11,'2026-08-18 23:46:05.846'),(132,1,12,'2026-08-18 23:46:05.848'),(133,1,13,'2026-08-18 23:46:05.851'),(134,1,14,'2026-08-18 23:46:05.855'),(135,1,15,'2026-08-18 23:46:05.859'),(136,1,16,'2026-08-18 23:46:05.862'),(137,1,17,'2026-08-18 23:46:05.865'),(138,1,18,'2026-08-18 23:46:05.868'),(139,1,19,'2026-08-18 23:46:05.871'),(140,1,20,'2026-08-18 23:46:05.873'),(141,1,21,'2026-08-18 23:46:05.876'),(142,1,22,'2026-08-18 23:46:05.878'),(143,1,23,'2026-08-18 23:46:05.881'),(144,1,24,'2026-08-18 23:46:05.884'),(145,1,25,'2026-08-18 23:46:05.886'),(146,1,26,'2026-08-18 23:46:05.889'),(147,1,27,'2026-08-18 23:46:05.892'),(148,1,28,'2026-08-18 23:46:05.894'),(149,1,29,'2026-08-18 23:46:05.897'),(150,1,30,'2026-08-18 23:46:05.900'),(151,1,31,'2026-08-18 23:46:05.903'),(152,1,32,'2026-08-18 23:46:05.906'),(153,1,33,'2026-08-18 23:46:05.908'),(154,1,34,'2026-08-18 23:46:05.911'),(155,1,35,'2026-08-18 23:46:05.914'),(156,1,36,'2026-08-18 23:46:05.917'),(157,1,37,'2026-08-18 23:46:05.919'),(158,1,38,'2026-08-18 23:46:05.922'),(159,1,39,'2026-08-18 23:46:05.925'),(160,1,40,'2026-08-18 23:46:05.928'),(161,1,41,'2026-08-18 23:46:05.932'),(162,1,42,'2026-08-18 23:46:05.935'),(163,1,43,'2026-08-18 23:46:05.938'),(164,1,44,'2026-08-18 23:46:05.940'),(165,1,45,'2026-08-18 23:46:05.943'),(166,1,46,'2026-08-18 23:46:05.947'),(167,1,47,'2026-08-18 23:46:05.949'),(168,1,48,'2026-08-18 23:46:05.952'),(169,1,49,'2026-08-18 23:46:05.954'),(170,1,50,'2026-08-18 23:46:05.957'),(171,1,51,'2026-08-18 23:46:05.960'),(172,1,52,'2026-08-18 23:46:05.964'),(173,1,53,'2026-08-18 23:46:05.967'),(174,1,54,'2026-08-18 23:46:05.971'),(175,1,55,'2026-08-18 23:46:05.975'),(176,1,56,'2026-08-18 23:46:05.978'),(177,1,57,'2026-08-18 23:46:05.981'),(178,1,58,'2026-08-18 23:46:05.984'),(179,1,59,'2026-08-18 23:46:05.988'),(180,1,60,'2026-08-18 23:46:05.991'),(181,1,61,'2026-08-18 23:46:05.994'),(182,1,62,'2026-08-18 23:46:05.997'),(183,1,63,'2026-08-18 23:46:06.000'),(184,1,64,'2026-08-18 23:46:06.003'),(185,1,65,'2026-08-18 23:46:06.007'),(186,1,66,'2026-08-18 23:46:06.010'),(187,1,67,'2026-08-18 23:46:06.013'),(188,1,68,'2026-08-18 23:46:06.016'),(189,1,69,'2026-08-18 23:46:06.019'),(190,2,1,'2026-08-18 23:46:06.022'),(191,2,2,'2026-08-18 23:46:06.025'),(192,2,3,'2026-08-18 23:46:06.028'),(193,2,4,'2026-08-18 23:46:06.032'),(194,2,5,'2026-08-18 23:46:06.035'),(195,2,6,'2026-08-18 23:46:06.039'),(196,2,7,'2026-08-18 23:46:06.041'),(197,2,8,'2026-08-18 23:46:06.045'),(198,2,9,'2026-08-18 23:46:06.049'),(199,2,10,'2026-08-18 23:46:06.051'),(200,2,11,'2026-08-18 23:46:06.055'),(201,2,12,'2026-08-18 23:46:06.058'),(202,2,13,'2026-08-18 23:46:06.061'),(203,2,14,'2026-08-18 23:46:06.064'),(204,2,15,'2026-08-18 23:46:06.067'),(205,2,16,'2026-08-18 23:46:06.070'),(206,2,17,'2026-08-18 23:46:06.073'),(207,2,18,'2026-08-18 23:46:06.076'),(208,2,19,'2026-08-18 23:46:06.079'),(209,2,20,'2026-08-18 23:46:06.082'),(210,2,21,'2026-08-18 23:46:06.085'),(211,2,22,'2026-08-18 23:46:06.088'),(212,2,23,'2026-08-18 23:46:06.091'),(213,2,24,'2026-08-18 23:46:06.093'),(214,2,25,'2026-08-18 23:46:06.097'),(215,2,26,'2026-08-18 23:46:06.100'),(216,2,27,'2026-08-18 23:46:06.103'),(217,2,28,'2026-08-18 23:46:06.107'),(218,2,29,'2026-08-18 23:46:06.110'),(219,2,30,'2026-08-18 23:46:06.113'),(220,2,31,'2026-08-18 23:46:06.116'),(221,2,32,'2026-08-18 23:46:06.119'),(222,2,33,'2026-08-18 23:46:06.121'),(223,2,34,'2026-08-18 23:46:06.124'),(224,2,35,'2026-08-18 23:46:06.127'),(225,2,36,'2026-08-18 23:46:06.129'),(226,2,37,'2026-08-18 23:46:06.133'),(227,2,38,'2026-08-18 23:46:06.136'),(228,2,39,'2026-08-18 23:46:06.139'),(229,2,40,'2026-08-18 23:46:06.142'),(230,2,41,'2026-08-18 23:46:06.145'),(231,2,42,'2026-08-18 23:46:06.147'),(232,2,43,'2026-08-18 23:46:06.150'),(233,2,44,'2026-08-18 23:46:06.152'),(234,2,45,'2026-08-18 23:46:06.155'),(235,2,46,'2026-08-18 23:46:06.157'),(236,2,47,'2026-08-18 23:46:06.161'),(237,2,48,'2026-08-18 23:46:06.164'),(238,2,49,'2026-08-18 23:46:06.166'),(239,2,50,'2026-08-18 23:46:06.169'),(240,2,51,'2026-08-18 23:46:06.172'),(241,2,52,'2026-08-18 23:46:06.174'),(242,2,53,'2026-08-18 23:46:06.177'),(243,2,54,'2026-08-18 23:46:06.180'),(244,2,55,'2026-08-18 23:46:06.182'),(245,2,56,'2026-08-18 23:46:06.185'),(246,2,58,'2026-08-18 23:46:06.187'),(247,2,59,'2026-08-18 23:46:06.190'),(248,2,60,'2026-08-18 23:46:06.192'),(249,2,61,'2026-08-18 23:46:06.195'),(250,2,62,'2026-08-18 23:46:06.197'),(251,2,63,'2026-08-18 23:46:06.200'),(252,2,64,'2026-08-18 23:46:06.202'),(253,2,65,'2026-08-18 23:46:06.204'),(254,2,66,'2026-08-18 23:46:06.207'),(255,2,67,'2026-08-18 23:46:06.210'),(256,2,68,'2026-08-18 23:46:06.213'),(257,1,70,'2026-08-19 02:41:33.719'),(258,1,71,'2026-08-19 02:41:33.724'),(259,1,72,'2026-08-19 02:41:33.728'),(260,1,73,'2026-08-19 02:41:33.731'),(261,1,74,'2026-08-19 02:41:33.736'),(262,1,75,'2026-08-19 02:41:33.740'),(263,2,70,'2026-08-19 02:41:33.780'),(264,2,71,'2026-08-19 02:41:33.785'),(265,2,72,'2026-08-19 02:41:33.788'),(266,2,73,'2026-08-19 02:41:33.791'),(267,2,74,'2026-08-19 02:41:33.794'),(268,2,75,'2026-08-19 02:41:33.798'),(269,3,72,'2026-08-19 02:41:33.833'),(270,3,73,'2026-08-19 02:41:33.838'),(271,3,75,'2026-08-19 02:41:33.842');
/*!40000 ALTER TABLE `role_permissions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `roles`
--

DROP TABLE IF EXISTS `roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(64) NOT NULL,
  `name` varchar(64) NOT NULL,
  `description` varchar(255) NOT NULL DEFAULT '',
  `is_system` tinyint(1) NOT NULL DEFAULT '0' COMMENT '1=系统内置角色，不可删除',
  `scope` varchar(16) NOT NULL DEFAULT 'global' COMMENT 'global/org/project',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_roles_code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='角色';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `roles`
--

LOCK TABLES `roles` WRITE;
/*!40000 ALTER TABLE `roles` DISABLE KEYS */;
INSERT INTO `roles` VALUES (1,'superadmin','超级管理员','平台全局管理员，拥有全部权限',1,'global','2026-08-18 23:46:05.188','2026-08-18 23:46:05.188',NULL),(2,'admin','管理员','组织管理员（用户管理/配额/计费）',1,'global','2026-08-18 23:46:05.190','2026-08-18 23:46:05.190',NULL),(3,'developer','开发者','项目内全量开发权限',1,'global','2026-08-18 23:46:05.193','2026-08-18 23:46:05.193',NULL),(4,'operator','运维','实例运维操作与流水线执行',1,'global','2026-08-18 23:46:05.195','2026-08-18 23:46:05.195',NULL),(5,'user','普通用户','自服务用户（仅自己资源）',1,'global','2026-08-18 23:46:05.198','2026-08-18 23:46:05.198',NULL),(6,'viewer','只读','全局只读',1,'global','2026-08-18 23:46:05.201','2026-08-18 23:46:05.201',NULL);
/*!40000 ALTER TABLE `roles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `schema_migrations`
--

DROP TABLE IF EXISTS `schema_migrations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `schema_migrations` (
  `version` varchar(191) NOT NULL,
  `applied_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `schema_migrations`
--

LOCK TABLES `schema_migrations` WRITE;
/*!40000 ALTER TABLE `schema_migrations` DISABLE KEYS */;
INSERT INTO `schema_migrations` VALUES ('000001_init.sql','2026-08-18 22:42:44.567'),('000002_audit_logs.sql','2026-08-18 23:46:04.904'),('000003_ecs.sql','2026-08-19 00:12:10.893'),('000004_infra.sql','2026-08-19 00:47:13.628'),('000005_apps.sql','2026-08-19 01:05:51.728'),('000006_pipeline.sql','2026-08-19 01:23:44.451'),('000007_webhooks.sql','2026-08-19 01:33:29.697'),('000008_ops.sql','2026-08-19 01:54:30.643'),('000009_org_quota_usage.sql','2026-08-19 02:16:46.046'),('000010_security.sql','2026-08-19 02:41:33.630');
/*!40000 ALTER TABLE `schema_migrations` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `secrets`
--

DROP TABLE IF EXISTS `secrets`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `secrets` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `org_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '0=全局/默认空间，其余=组织内',
  `name` varchar(128) NOT NULL,
  `value_cipher` text NOT NULL COMMENT 'AES-256-GCM 密文（base64，密钥由 JWT_SECRET 派生）',
  `created_by` bigint unsigned DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_secrets_org_name` (`org_id`,`name`,`deleted_at`),
  KEY `idx_secrets_org` (`org_id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='密钥托管（加密存储）';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `secrets`
--

LOCK TABLES `secrets` WRITE;
/*!40000 ALTER TABLE `secrets` DISABLE KEYS */;
INSERT INTO `secrets` VALUES (1,3,'DB_PASSWORD_P11','939hTzy6yebIeb1FLXYueiitm6j2cOLr1x3DaECY5dh5qvHSEpnoh0xXtI4=',1,'2026-08-19 02:43:14.722','2026-08-19 02:43:14.810','2026-08-19 02:43:14.810'),(2,3,'DB_PASSWORD_P11','C15Ks0lKjBG5bCl6RxF9ABvFbwzq7J1tocGey6WlSODdXG1dfec8J3C5mm8=',1,'2026-08-19 02:44:32.071','2026-08-19 02:44:32.154','2026-08-19 02:44:32.154'),(3,3,'DB_PASSWORD_P11','X2VNcVusaSEwGie3rB8hsM5I+wS/SdAiXBHvfTY+yS/QmvrT4GWf7qpEbBI=',1,'2026-08-19 02:45:48.696','2026-08-19 02:45:48.772','2026-08-19 02:45:48.772');
/*!40000 ALTER TABLE `secrets` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `security_reports`
--

DROP TABLE IF EXISTS `security_reports`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `security_reports` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `kind` varchar(32) NOT NULL COMMENT 'baseline/image/secret',
  `score` int NOT NULL DEFAULT '100',
  `finding_count` int NOT NULL DEFAULT '0',
  `summary` text COMMENT '发现项 JSON 数组',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_security_reports_kind` (`kind`,`id`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='安全扫描报告';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `security_reports`
--

LOCK TABLES `security_reports` WRITE;
/*!40000 ALTER TABLE `security_reports` DISABLE KEYS */;
INSERT INTO `security_reports` VALUES (1,'baseline',65,18,'[{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-i-f2769eac89fcc103\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-i-79258b28f6e449cf\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-i-f82fbf823e70dee8\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-i-8868a5136c00b7d3\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-app-3-v14\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-app-3-v15\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-app-3-v13\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"cloud-api\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"cloud-proxy\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"cloud-web\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"dx-registry\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"dx-redis\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"dx-mysql\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"mirofish\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"zen_poitras\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"bettafish-db\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"bettafish\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"xenodochial_kilby\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"}]','2026-08-19 02:43:14.548'),(2,'image',66,8,'[{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"dxcloud-backend:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"dxcloud-frontend:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"busybox:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"hello-world:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"ghcr.io/666ghj/mirofish:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"low\",\"kind\":\"image\",\"target\":\"sha256:d2364\",\"message\":\"镜像体积 13.0 GB（建议瘦身）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"ghcr.io/666ghj/bettafish:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"low\",\"kind\":\"image\",\"target\":\"sha256:38939\",\"message\":\"镜像体积 14.7 GB（建议瘦身）\"}]','2026-08-19 02:43:14.660'),(3,'baseline',65,18,'[{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-i-f2769eac89fcc103\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-i-79258b28f6e449cf\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-i-f82fbf823e70dee8\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-i-8868a5136c00b7d3\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-app-3-v14\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-app-3-v15\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-app-3-v13\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"cloud-api\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"cloud-proxy\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"cloud-web\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"dx-registry\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"dx-redis\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"dx-mysql\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"mirofish\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"zen_poitras\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"bettafish-db\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"bettafish\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"xenodochial_kilby\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"}]','2026-08-19 02:44:31.919'),(4,'image',66,8,'[{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"dxcloud-backend:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"dxcloud-frontend:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"busybox:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"hello-world:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"ghcr.io/666ghj/mirofish:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"low\",\"kind\":\"image\",\"target\":\"sha256:d2364\",\"message\":\"镜像体积 13.0 GB（建议瘦身）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"ghcr.io/666ghj/bettafish:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"low\",\"kind\":\"image\",\"target\":\"sha256:38939\",\"message\":\"镜像体积 14.7 GB（建议瘦身）\"}]','2026-08-19 02:44:32.023'),(5,'baseline',65,18,'[{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-i-f2769eac89fcc103\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-i-79258b28f6e449cf\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-i-f82fbf823e70dee8\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-i-8868a5136c00b7d3\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-app-3-v14\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-app-3-v15\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"medium\",\"kind\":\"baseline\",\"target\":\"dx-app-3-v13\",\"message\":\"以 root 运行（建议指定非 root 用户）\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"cloud-api\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"cloud-proxy\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"cloud-web\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"dx-registry\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"dx-redis\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"dx-mysql\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"mirofish\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"zen_poitras\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"bettafish-db\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"bettafish\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"},{\"severity\":\"info\",\"kind\":\"baseline\",\"target\":\"xenodochial_kilby\",\"message\":\"非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查\"}]','2026-08-19 02:45:48.531'),(6,'image',66,8,'[{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"dxcloud-backend:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"dxcloud-frontend:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"busybox:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"hello-world:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"ghcr.io/666ghj/mirofish:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"low\",\"kind\":\"image\",\"target\":\"sha256:d2364\",\"message\":\"镜像体积 13.0 GB（建议瘦身）\"},{\"severity\":\"medium\",\"kind\":\"image\",\"target\":\"ghcr.io/666ghj/bettafish:latest\",\"message\":\"使用 latest 标签（生产环境应使用不可变版本号）\"},{\"severity\":\"low\",\"kind\":\"image\",\"target\":\"sha256:38939\",\"message\":\"镜像体积 14.7 GB（建议瘦身）\"}]','2026-08-19 02:45:48.645');
/*!40000 ALTER TABLE `security_reports` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `system_settings`
--

DROP TABLE IF EXISTS `system_settings`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `system_settings` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `key` varchar(128) NOT NULL,
  `value` json DEFAULT NULL,
  `description` varchar(255) NOT NULL DEFAULT '',
  `updated_by` bigint unsigned DEFAULT NULL,
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_system_settings_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='系统设置（配额默认值/安全策略/保留期等）';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `system_settings`
--

LOCK TABLES `system_settings` WRITE;
/*!40000 ALTER TABLE `system_settings` DISABLE KEYS */;
/*!40000 ALTER TABLE `system_settings` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `user_roles`
--

DROP TABLE IF EXISTS `user_roles`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL,
  `role_id` bigint unsigned NOT NULL,
  `org_id` bigint unsigned DEFAULT NULL COMMENT '组织级授权时填写',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_roles` (`user_id`,`role_id`,`org_id`),
  KEY `idx_user_roles_org` (`org_id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户-角色';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `user_roles`
--

LOCK TABLES `user_roles` WRITE;
/*!40000 ALTER TABLE `user_roles` DISABLE KEYS */;
INSERT INTO `user_roles` VALUES (1,1,1,NULL,'2026-08-18 23:46:06.267'),(2,2,5,NULL,'2026-08-18 23:46:51.607'),(3,3,5,NULL,'2026-08-19 02:29:50.951'),(4,4,5,NULL,'2026-08-19 02:33:11.011'),(5,5,5,NULL,'2026-08-19 02:43:14.893'),(6,6,5,NULL,'2026-08-19 02:44:32.241'),(7,7,5,NULL,'2026-08-19 02:46:53.873');
/*!40000 ALTER TABLE `user_roles` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(64) NOT NULL,
  `email` varchar(128) NOT NULL,
  `password_hash` varchar(255) NOT NULL DEFAULT '',
  `nickname` varchar(64) NOT NULL DEFAULT '',
  `avatar_url` varchar(512) NOT NULL DEFAULT '',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '1=active 2=disabled 3=locked',
  `last_login_at` datetime(3) DEFAULT NULL,
  `last_login_ip` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_username` (`username`),
  UNIQUE KEY `uk_users_email` (`email`),
  KEY `idx_users_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='平台用户';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES (1,'admin','admin@dxcloud.local','$2a$10$AzKOn95alju4Vx4vD.vYeergZdHob.Ma.EKOxxVY16w4zd/G0fQK2','Administrator','',1,'2026-08-19 03:07:37.722','172.21.0.1','2026-08-18 23:46:06.264','2026-08-19 03:07:37.723',NULL),(2,'alice','alice@test.com','$2a$10$V40JL1IYNL7fCoEVQN1qneo.TvfGCXnuq/482Q7GUKrmkZSjiR13y','alice','',1,'2026-08-19 00:36:07.046','172.21.0.1','2026-08-18 23:46:51.602','2026-08-19 00:36:07.046',NULL),(3,'bob0819022950','bob0819022950@dx.dev','$2a$10$eFa5lPv/NJr68CPEtM.VwOwLsipAma0u.RivC63T6AGlzR.YKunYW','bob0819022950','',1,'2026-08-19 02:29:51.051','172.21.0.1','2026-08-19 02:29:50.943','2026-08-19 02:29:51.051',NULL),(4,'bob0819023310','bob0819023310@dx.dev','$2a$10$e5hKqerkgy3YKtI.fVZZCOESvmw4dIn0hYk965LWgRKjiZiopiLxm','bob0819023310','',1,'2026-08-19 02:33:11.077','172.21.0.1','2026-08-19 02:33:11.006','2026-08-19 02:33:11.078',NULL),(5,'locku0819024314','locku0819024314@dx.dev','$2a$10$kNQEpTXPPixxLo2EWruQd.sY7rX9Mr89wVRPC3xPINSIYdhqhFDL2','locku0819024314','',1,'2026-08-19 02:44:17.508','172.21.0.1','2026-08-19 02:43:14.888','2026-08-19 02:44:17.508',NULL),(6,'locku0819024432','locku0819024432@dx.dev','$2a$10$b73sX8.2YMy/YfkIUEA22ODyjkRlPT.GQ8SfKyADT48SVmcP28ztW','locku0819024432','',1,'2026-08-19 02:45:34.796','172.21.0.1','2026-08-19 02:44:32.236','2026-08-19 02:45:34.797',NULL),(7,'locku0819024653','locku0819024653@dx.dev','$2a$10$zsOHCme4rgKnsYONqknXv.CGix14Fzsa50xg/7lDN4iPKqr8WkJc.','locku0819024653','',1,'2026-08-19 02:47:56.452','172.21.0.1','2026-08-19 02:46:53.865','2026-08-19 02:47:56.452',NULL);
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `webhooks`
--

DROP TABLE IF EXISTS `webhooks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `webhooks` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `project_id` bigint unsigned DEFAULT NULL,
  `pipeline_id` bigint unsigned NOT NULL,
  `provider` varchar(16) NOT NULL COMMENT 'github/gitlab/gitee',
  `secret_enc` varchar(512) NOT NULL DEFAULT '' COMMENT 'AES-GCM 加密的签名密钥',
  `branch_filter` varchar(128) NOT NULL DEFAULT '' COMMENT '如 main 或 release/*（空=全部）',
  `events` varchar(64) NOT NULL DEFAULT 'push',
  `status` tinyint NOT NULL DEFAULT '1',
  `hook_code` varchar(32) NOT NULL COMMENT 'URL 随机码',
  `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_webhooks_code` (`hook_code`),
  KEY `idx_webhooks_pipeline` (`pipeline_id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Git Webhook 注册';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `webhooks`
--

LOCK TABLES `webhooks` WRITE;
/*!40000 ALTER TABLE `webhooks` DISABLE KEYS */;
INSERT INTO `webhooks` VALUES (1,NULL,4,'github','eFeIzOL5GRVD62Dqyw/GwEX16msbWlFW38ts9ZFtoAXEh6EOYi6lTpUcGQ==','main','push',1,'aea857b1dad67536','2026-08-19 01:34:18.016','2026-08-19 01:34:18.016',NULL),(2,NULL,4,'github','BNMWHywsnEkmleDwoBbJhUPbljDmkHA5XB9oe9qxFB24yQWkuIuusY8qpw==','release/*','push',1,'d42811c5227ca84f','2026-08-19 01:50:10.111','2026-08-19 01:50:10.168','2026-08-19 01:50:10.168');
/*!40000 ALTER TABLE `webhooks` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Dumping routines for database 'dxcloud'
--
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-08-19  3:07:42
