# Finances - Product Requirements Document (PRD)

## 1. Executive Summary

### 1.1 Project Overview

**Project Name:** Finances
**Version:** 1.0
**Date:** August 15, 2025
**Document Owner:** Matheus Lopes
**Status:** Draft

### 1.2 Purpose

During the dificult moments everyone should have the hability to manage its finance. Individuals must make
the best decisions so that they don't run out of money, and also being able to multiply their wallet.

There are a lot of things that can explain the current finance situation of an individual. The way that
people spend money tells much about them, and the objective of this project is to track the expenses of
each individual and provide insights so that they can identify the good or bad financial decisions they are
making.

The first step to organize a finance strategy is by having data to generate information about an individual.
This project has the primary objective of collecting all the finance transactions that an individual does, by
letting the user upload CSV files of their bank movements and ingest the data to the system to analise the
information, and based on the data provided, the system can generate finance charts and insights for users
so that they can have this better understanding of their financial life.

This is very important because by having this knowledge, users can make smart moves and achieve an financial
independency.

### 1.3 Scope

**Included in Scope:**

- Allow users to upload CSV files of their bank transactions.
- Ingest and process transaction data for analysis.
- Generate financial charts and insights based on user data.
- Provide a dashboard for users to view and track their expenses.
- Support for multiple bank formats for CSV uploads.
- Provide a page where users can upload their csv data.

**Excluded from Scope:**

- Direct integration with banking APIs for automatic data import.
- Investment portfolio management features.
- Bill payment or money transfer functionalities.
- Tax calculation or filing services.

## 2. Product Vision & Goals

### 2.1 Vision Statement

The goal is to help users achieve the financial wealth and being able to be stable in their lives.
We believe that every individual should enjoy the life and having the power of control their
destiny. This tool would help them manage their finance by providing the necessary feedback
around their transactions activity.

### 2.2 Business Objectives

- **Primary Goal:** Help users manage personal finances effectively
- **Secondary Goal:**
  - Provide visibility on the user's expenses

### 2.3 Success Metrics

- User adoption rate
- User engagement metrics
- Business value metrics

## 3. Product Features & Requirements

### 3.1 Core Features (MVP)

#### 3.1.1 Feature 1: Create user

- **Description:** To sign in and use the application features real persons should be able to
  create user accounts.
- **User Story:** As a user, I want to be able to create my account so that I can interact
  with the private modules of the finance application.
- **Acceptance Criteria:**
  - User must provide an email, or sign in with Google.
  - A user record must have the person's first name, last name, and optionally a profile picture.
  - We should not expect the user to provide a password
- **Priority:** High
- **Effort Estimate:** 2 Hours

#### 3.1.2 Feature 2: Sign in

- **Description:** Anyone who created an user, should be able to sign in to the application.
  The main sign in method should be made with passwordless authentication.
- **User Story:** As a user, I need to sign in to the application by providing my user email
  and receiving a temporary code that should be valid through 10 minutes. I should be able to
  inform this code received in my email to login to the application. If I do not inform the code
  after 10 minutes, the application should invalidade the code and I should not be able to sign
  in using this code anymore.
- **Acceptance Criteria:**
  - Providing a valid user email, should send an email message to the user with a signing code.
    - Not receiving the code would break the sign in flow.
    - Code should be valid through 10 minutes.
    - Code should be invalidated after 10 minutes.
    - Users can't sign in using an invalid code.
  - Providing the authentication code received by email, should authenticate the user.
    - Not authenticating user to the application when code is provided, should fail.
    - Sending an invalid code, should return a descriptive error message.
- **Priority:** High
- **Effort Estimate:** 5 Hours

#### 3.1.3 Feature 3: Send email notifications

- **Description:** We should keep one principal channel of communication with our users and
  we are going to keep email communications for it.
- **User Story:** As a user, when performing sign in operations, the application should send
  me the authentication code through email.
- **Acceptance Criteria:**
  - Service to provide an API to send emails to users.
  - API should specify the available messages and which parameters the user should provide.
  - Templates to make available:
    - Authentication Code.
    - Welcome message
- **Priority:** High
- **Effort Estimate:** 5 Hours

#### 3.1.4 Feature 4: Uploading Transactions CSV

- **Description:** The system should provide an API for uploading the bank transactions CSV
  files.
- **User Story:** As a user, I should expect the system to provide me a resource to upload my
  transactions CSV files from different bank sources.
- **Acceptance Criteria:**
  - CSV file to be storaged in AWS S3.
- **Priority:** High
- **Effort Estimate:** 5 Hours

#### 3.1.5 Feature 5: Read CSV Transactions

- **Description:** Once the CSV files get available in S3 bucket, the system should be able
  to read an S3 upload event and perform a read operation to the CSV file. This should create
  transaction entities and store each transaction record to the database.
- **User Story:** As a system, I should be listen to objects being dropped on my S3 bucket
  to perform a read operation on the CSV file and create transaction records on my DB.
- **Acceptance Criteria:**
  - Should read all the transactions and store the values in a database
- **Priority:** High
- **Effort Estimate:** 5 Hours

#### 3.1.6 Feature 6: List transactions

- **Description:** Provide an API to list all the transactions, that accepts the following
  query parameter:
  - **order_by:** query parameter that would indicate which field to apply the order logic.
    - Default value: transaction date time
  - **sort:** Indicate how the order sorting would behave.
    - Options: [ASC, DESC]
    - Default value: DESC
  - **TODO**: [Parei_aqui_pensando_em_outros_parametros]
- **User Story:**
- **Acceptance Criteria:**
- **Priority:** [High|Medium|Low]
- **Effort Estimate:** X Hours

## 4. System Architecture

### 4.1 High-Level Architecture

```
[Frontend] <-> [API Gateway] <-> [Backend Services] <-> [Database]
```

### 4.2 Technology Stack

#### 4.2.1 Frontend

- Framework: RemixJS
- State Management: UI state will be handled by the Framework Server Side rendering hydration strategy.
- UI Library: To be defined.

#### 4.2.2 Backend

- Runtime: Golang
- Framework: Go Chi, and more libraries.
- Database: Postgres
- Cache: Redis

#### 4.2.3 Infrastructure

- Cloud Provider: AWS
- Container: Docker
- CI/CD: GitHub Actions

### 4.3 Data Model

- Entity Relationship Diagram
- User <-> Transactions

### 5 Monitoring & Logging

Framework: OTEL
UI: To be defined.

---

**Document History**

| Version | Date       | Author        | Changes              |
| ------- | ---------- | ------------- | -------------------- |
| 1.0     | 2025-08-15 | Matheus Lopes | Initial PRD template |

---

_This document is a living document and will be updated as the project evolves._
