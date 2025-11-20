# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2024-11-20

### Added
- Complete rewrite with thread-safe implementation
- YAML-based configuration with environment variable support
- Structured logging with slog
- Graceful shutdown handling with context
- Connection pooling for Kafka producer
- Worker pool support in scheduler
- Comprehensive test coverage
- Docker and docker-compose support
- CI/CD pipeline with GitHub Actions
- Makefile for build automation
- Detailed README with examples
- Configuration validation
- Statistics tracking for scheduler
- Template functions for UUID, GUID, timestamps, random numbers
- Support for batch message sending
- Comprehensive error handling

### Changed
- Migrated from JSON to YAML configuration
- Replaced custom scheduler with robust implementation
- Switched from DialLeader to kafka.Writer for better performance
- Improved template generation with crypto/rand for security
- Refactored project structure (cmd/internal pattern)
- Updated all dependencies to latest versions

### Fixed
- Race conditions in scheduler and template generation
- Memory leaks in Kafka connections
- Improper error handling throughout the codebase
- Missing graceful shutdown
- Thread-unsafe random number generation

### Removed
- Legacy pkg structure
- Unused dependencies (beevik/guid, robfig/cron, gods)
- Overcomplicated scheduler implementation
- JSON configuration support

## [1.0.0] - Initial Release

### Added
- Basic Kafka message sending
- Template-based payload generation
- Simple scheduler
- JSON configuration
