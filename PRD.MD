# Product Requirements Document

## Product name

Working name: **Approval Inbox**

## Document status

Draft MVP PRD

## Product summary

Approval Inbox is a centralised human-approval platform for AI agents and automated workflows.

It allows companies to:

* receive proposed actions from systems such as n8n, LangGraph and custom applications;
* evaluate those actions against central policies;
* route sensitive actions to the correct human approver;
* collect approve, edit or reject decisions;
* return the decision to the original workflow so it can continue;
* maintain a complete audit history.

Approval Inbox does not execute the final business action.

The original workflow remains responsible for actions such as:

* issuing a refund;
* sending an email;
* publishing content;
* updating a CRM;
* deleting records;
* changing supplier details.

The product acts as the central policy and approval layer between an automated workflow proposing an action and that workflow continuing its execution.

---

# 1. Problem statement

As companies deploy more AI agents and automated workflows, actions requiring human judgement become fragmented across:

* email;
* Slack;
* Microsoft Teams;
* workflow dashboards;
* CRM notifications;
* internal applications;
* custom approval interfaces.

Each team may implement approval logic differently.

For example:

* one n8n workflow sends a Slack message;
* another workflow waits for an email response;
* a LangGraph agent uses a custom human-in-the-loop interface;
* another internal application stores approvals in a database;
* policies such as refund thresholds are hard-coded separately in several workflows.

This creates several problems:

* approvers must monitor multiple systems;
* approval rules become inconsistent;
* developers repeatedly rebuild similar approval infrastructure;
* approval requests may lack sufficient evidence;
* there is no central audit history;
* workflow continuation is unreliable or difficult to maintain;
* companies struggle to safely give AI agents permission to act.

---

# 2. Product vision

Provide one central checkpoint between machine-proposed actions and business impact.

The product should enable companies to safely increase automation by allowing routine actions to continue automatically while routing exceptional or higher-risk actions to the correct human.

Long-term vision:

> Every consequential action proposed by an AI agent can be evaluated against company policy before the original workflow continues.

---

# 3. MVP goal

The MVP must prove that teams running production workflows want a reusable external service for:

* central approval policies;
* one approval inbox;
* consistent decision context;
* reliable workflow continuation;
* central audit history.

The MVP should focus on n8n and generic webhook integrations first, with LangGraph as the second supported integration.

---

# 4. Core value proposition

> Define approval rules once, review risky AI actions in one inbox and return the decision so the original workflow can continue.

For developers:

> Stop rebuilding approval logic for every workflow.

For operations teams:

> Review all exceptional AI actions from one place.

For company leadership:

> Deploy more automation without giving AI agents unrestricted authority.

---

# 5. Target users

## Primary customer profile

Small and medium-sized companies or automation consultancies that:

* operate several automated or AI-assisted workflows;
* use n8n, LangGraph or custom applications;
* automate customer support, finance or operational processes;
* need human review for higher-risk actions;
* do not have a mature enterprise governance platform.

Recommended initial company size:

* 20 to 500 employees.

## Primary personas

### Automation developer

Responsibilities:

* builds and maintains workflows;
* connects n8n, LangGraph or custom services;
* manages workflow pause and resume logic.

Needs:

* simple API;
* reusable integration templates;
* reliable callbacks;
* clear documentation;
* minimal custom approval code.

### Operations administrator

Responsibilities:

* defines approval policies;
* assigns approver groups;
* monitors pending and failed requests.

Needs:

* simple policy builder;
* central visibility;
* escalation settings;
* audit history.

### Human approver

Responsibilities:

* reviews proposed actions;
* evaluates evidence;
* approves, edits or rejects actions.

Needs:

* clear action summary;
* supporting evidence;
* policy explanation;
* fast and simple decision process.

### Compliance or security reviewer

Responsibilities:

* reviews approval records;
* verifies that policies are followed;
* investigates incidents.

Needs:

* immutable audit trail;
* approver identity;
* policy snapshots;
* decision timestamps;
* exportable records.

---

# 6. Initial use cases

## Primary MVP use case

Customer refund approval.

Example:

An n8n workflow proposes a £75 refund.

Company policy states:

* refunds up to £30 are automatically allowed;
* refunds above £30 require a support manager;
* refunds above £500 are blocked or require two approvers.

The workflow sends the proposed action to Approval Inbox.

Approval Inbox evaluates the policy and creates a pending request.

The manager reviews:

* customer;
* order;
* refund amount;
* reason;
* evidence;
* previous refunds;
* affected systems;
* reversibility.

The manager approves, changes the amount or rejects the request.

Approval Inbox sends a signed decision back to n8n.

n8n resumes and performs the refund or stops.

## Secondary use cases

* sending quotations with discounts above a threshold;
* changing supplier bank details;
* publishing AI-generated content;
* deleting customer records;
* issuing account credits;
* sending sensitive customer communications;
* approving high-cost travel bookings;
* changing user permissions.

---

# 7. Product scope

## In scope for MVP

* user authentication;
* workspaces;
* team invitations;
* basic roles;
* approval-request API;
* common action schema;
* simple policy builder;
* policy evaluation;
* central approval inbox;
* request detail page;
* approve;
* edit and approve;
* reject;
* comments;
* expiration;
* basic escalation;
* outgoing decision webhooks;
* signed decision payloads;
* idempotency;
* audit log;
* n8n workflow template;
* generic webhook integration;
* basic LangGraph helper;
* Slack notifications;
* request and delivery status.

## Out of scope for MVP

* direct execution of business actions;
* storing customer Stripe, CRM or ERP credentials;
* visual workflow builder;
* full enterprise policy language;
* native Microsoft Teams integration;
* mobile application;
* advanced analytics;
* AI-generated policy recommendations;
* MCP integration;
* marketplace of connectors;
* SAML or enterprise SSO;
* on-premise deployment;
* regulatory compliance certification;
* rollback or transaction compensation;
* direct database execution;
* complex multi-stage enterprise approval chains.

---

# 8. Product principles

## Approval only

Approval Inbox decides whether an action is:

* allowed;
* pending approval;
* approved;
* approved with changes;
* rejected;
* blocked;
* expired.

The source workflow owns final execution.

## Platform independent

The central product must not depend on n8n or LangGraph internally.

Each platform connects through an adapter.

## Exact action binding

An approval must be tied to the exact action payload reviewed by the human.

## Decision-ready context

Approvers should see enough information to make a decision without opening several systems.

## Exceptions, not everything

The product should route only actions that genuinely need human judgement.

## Reliable continuation

A human decision and workflow continuation are separate events and must be tracked separately.

---

# 9. Core user journeys

## Journey 1: Administrator creates a workspace

1. User creates an account.
2. User creates a workspace.
3. User names the workspace.
4. User receives an API key.
5. User invites team members.
6. User assigns roles.
7. User connects the first workflow.

Expected result:

The workspace is ready to receive approval requests.

---

## Journey 2: Administrator connects n8n

1. Administrator opens Integrations.
2. Administrator selects n8n.
3. Product displays:

   * API endpoint;
   * API key;
   * webhook signing secret;
   * importable n8n template;
   * setup instructions.
4. Administrator imports the workflow template.
5. Administrator submits a test action.
6. Product displays the test request.
7. Administrator confirms the integration.

Expected result:

An n8n execution can submit an action, pause and resume after a human decision.

---

## Journey 3: Administrator creates a policy

1. Administrator selects Create Policy.
2. Administrator enters a policy name.
3. Administrator selects an action type.
4. Administrator adds optional conditions.
5. Administrator chooses the result:

   * allow;
   * require approval;
   * block.
6. If approval is required, administrator selects:

   * approver user;
   * approver group;
   * deadline;
   * escalation target.
7. Administrator activates the policy.

Example:

```text
WHEN
Action type equals Customer Refund

AND
Amount is greater than £30

THEN
Require approval from Support Managers

Expire after 2 hours
```

Expected result:

Incoming actions matching the conditions follow the configured policy.

---

## Journey 4: Workflow submits an action

1. Workflow prepares a proposed action.
2. Workflow calls the Approval Inbox API.
3. Product validates the request.
4. Product checks the idempotency key.
5. Product normalises and hashes the action.
6. Product evaluates active policies.
7. Product returns:

   * allow;
   * block;
   * pending approval.

Expected result:

The workflow knows whether to continue immediately, stop or wait.

---

## Journey 5: Approver reviews an action

1. Approver receives a Slack or web notification.
2. Approver opens the approval request.
3. Approver sees:

   * proposed action;
   * reason;
   * evidence;
   * affected systems;
   * before-and-after values;
   * policy triggered;
   * deadline;
   * reversible status.
4. Approver selects:

   * approve;
   * edit and approve;
   * reject.
5. Approver optionally adds a comment.

Expected result:

The decision is stored and queued for delivery to the source workflow.

---

## Journey 6: Workflow resumes

1. Approval Inbox creates a signed decision payload.
2. Approval Inbox delivers the decision to the workflow continuation target.
3. The workflow verifies the signature.
4. The workflow checks that the decision is not expired.
5. The workflow uses the approved action payload.
6. The workflow continues or stops.
7. The workflow optionally acknowledges receipt.

Expected result:

The original workflow continues according to the human decision.

---

# 10. Functional requirements

## 10.1 Authentication and workspace management

The system must allow users to:

* register;
* sign in;
* sign out;
* create a workspace;
* invite users;
* deactivate users;
* assign workspace roles.

MVP roles:

* Owner;
* Administrator;
* Approver;
* Viewer.

---

## 10.2 API keys

Administrators must be able to:

* create API keys;
* name API keys;
* view the key once;
* revoke API keys;
* see the last-used timestamp.

API keys must belong to one workspace.

---

## 10.3 Action submission API

Endpoint:

```text
POST /v1/approval-requests
```

Required request fields:

* idempotency key;
* action type;
* title;
* action parameters;
* source platform;
* source workflow identifier;
* source execution identifier;
* continuation strategy.

Optional fields:

* description;
* reason;
* evidence;
* confidence;
* reversible status;
* affected systems;
* deadline;
* metadata.

Example:

```json
{
  "action": {
    "type": "customer.refund",
    "title": "Refund £75 to customer",
    "parameters": {
      "order_id": "10482",
      "amount": 75,
      "currency": "GBP"
    }
  },
  "context": {
    "reason": "Customer provided evidence of damaged goods",
    "reversible": false
  },
  "source": {
    "platform": "n8n",
    "workflow_id": "wf_123",
    "execution_id": "exec_456"
  },
  "continuation": {
    "strategy": "webhook",
    "url": "https://..."
  }
}
```

---

## 10.4 Action schema

The product must maintain a common action model.

Minimum fields:

* action ID;
* action type;
* title;
* description;
* parameters;
* reason;
* evidence;
* source platform;
* source workflow;
* source execution;
* affected systems;
* reversibility;
* original action hash;
* created time.

The schema must allow unknown or customer-specific action types.

---

## 10.5 Policy engine

The policy engine must support:

### Conditions

* action type equals;
* source platform equals;
* target system equals;
* numeric field greater than;
* numeric field less than;
* string field equals;
* risk level equals;
* reversible equals true or false.

### Outcomes

* allow;
* require approval;
* block.

### Approval settings

* approver user;
* approver group;
* deadline;
* escalation target;
* prevent self-approval.

### Evaluation behaviour

* only active policies are evaluated;
* policies are evaluated by priority;
* the matched policy is stored as a snapshot;
* policy changes do not modify existing requests;
* if no policy matches, workspace default behaviour applies.

Recommended default behaviour for MVP:

```text
Require approval
```

A workspace may later choose:

```text
Allow by default
Block by default
```

---

## 10.6 Approval inbox

The inbox must show:

* request title;
* action type;
* source workflow;
* source platform;
* requested time;
* deadline;
* assigned approver;
* risk or impact indicator;
* current status.

Required filters:

* pending;
* assigned to me;
* approved;
* approved with changes;
* rejected;
* blocked;
* expired;
* all.

Optional filters:

* action type;
* platform;
* workflow;
* approver;
* policy.

---

## 10.7 Request detail page

The request detail page must display:

* proposed action;
* original parameters;
* reason;
* evidence;
* affected systems;
* reversible status;
* source platform;
* workflow name or identifier;
* policy triggered;
* assigned approver;
* deadline;
* activity history;
* comments.

Decision actions:

* approve;
* edit and approve;
* reject.

---

## 10.8 Edit and approve

The system must allow an approver to modify fields that the source integration marks as editable.

Example:

```text
Original refund amount: £75
Approved refund amount: £50
```

The system must store:

* original action;
* original action hash;
* approved action;
* approved action hash;
* fields changed;
* approver;
* comment;
* timestamp.

The original action must remain immutable.

---

## 10.9 Comments

Users must be able to add comments to pending and resolved requests.

Comments must record:

* author;
* body;
* timestamp.

Comments do not automatically change the workflow decision.

---

## 10.10 Decision delivery

Supported MVP continuation strategies:

* webhook callback;
* polling.

Later:

* direct LangGraph server adapter;
* event queues;
* native platform adapters.

Decision payload must include:

* request ID;
* decision ID;
* decision;
* original action hash;
* approved action;
* approved action hash;
* approver identity;
* matched policy;
* issue time;
* expiration time;
* signature.

---

## 10.11 Webhook delivery

The system must:

* send decisions asynchronously;
* retry failed requests;
* track attempt count;
* store response status;
* store response timestamp;
* allow manual retry;
* prevent duplicate decision creation.

Suggested retry schedule:

* immediately;
* 5 seconds;
* 30 seconds;
* 2 minutes;
* 10 minutes;
* 1 hour.

---

## 10.12 Polling

Connected workflows must be able to retrieve request status:

```text
GET /v1/approval-requests/{request_id}
```

Possible responses:

* pending;
* approved;
* approved_with_changes;
* rejected;
* blocked;
* expired;
* cancelled.

---

## 10.13 Workflow acknowledgement

Source workflows should optionally report:

* decision received;
* workflow resumed;
* workflow completed;
* workflow failed.

Endpoint:

```text
POST /v1/approval-requests/{request_id}/acknowledgements
```

The product must clearly distinguish:

* human decision status;
* decision-delivery status;
* source workflow acknowledgement;
* final business-action status.

The MVP must not claim that an action was executed unless the source workflow reports it.

---

## 10.14 Audit log

The audit log must record:

* request received;
* request validated;
* policy evaluation;
* policy matched;
* approver assignment;
* notification sent;
* request viewed;
* comment added;
* action edited;
* approved;
* rejected;
* expired;
* callback attempted;
* callback delivered;
* callback failed;
* acknowledgement received.

Audit records must be immutable.

---

## 10.15 Notifications

MVP notification channels:

* in-app;
* Slack;
* email optional.

Slack notification should include:

* request title;
* proposed action summary;
* policy triggered;
* deadline;
* link to request.

High-risk requests should require authenticated confirmation in the web application.

---

# 11. Integration requirements

## 11.1 Generic webhook integration

The integration must support:

1. Submit action.
2. Receive immediate result.
3. Pause if result is pending.
4. Receive signed webhook decision.
5. Continue or stop.
6. Optionally acknowledge the decision.

This is the base integration for all platforms.

---

## 11.2 n8n integration

The MVP must provide an importable n8n template containing:

* action preparation;
* HTTP Request node;
* Wait node;
* resume webhook;
* signature verification;
* decision switch;
* approve branch;
* edit branch;
* reject branch.

Expected flow:

```text
AI or business logic
        ↓
Prepare proposed action
        ↓
Submit to Approval Inbox
        ↓
Pending?
   ├── No → Continue or stop
   └── Yes → Wait
                ↓
         Receive decision
                ↓
         Verify signature
                ↓
          Continue workflow
```

The n8n template should demonstrate a customer refund use case.

---

## 11.3 LangGraph integration

The MVP should provide a basic Python helper or code example that supports:

* action submission;
* stable idempotency key;
* `interrupt()` when approval is pending;
* durable thread ID;
* decision resumption;
* approved action extraction;
* signature verification.

Expected flow:

```text
Agent proposes tool call
        ↓
Approval node
        ↓
Approval required
        ↓
interrupt()
        ↓
Human decision
        ↓
Command(resume=...)
        ↓
Tool execution
```

The customer’s LangGraph application should own:

* checkpointer;
* graph state;
* thread ID;
* tool credentials;
* tool execution.

---

# 12. Status model

## Approval status

```text
received
evaluating
allowed
blocked
pending
approved
approved_with_changes
rejected
expired
cancelled
```

## Delivery status

```text
not_required
pending
attempting
delivered
acknowledged
retrying
permanently_failed
```

Approval and delivery statuses must remain separate.

---

# 13. Policy examples

## Refund policy

```text
WHEN
Action type equals Customer Refund

AND
Amount is greater than £30

THEN
Require approval from Support Managers
```

## Supplier bank details

```text
WHEN
Action type equals Supplier Bank Details Change

THEN
Require approval from Finance Manager

AND
Prevent self-approval
```

## Record deletion

```text
WHEN
Action type equals Delete CRM Records

AND
Record count is greater than 100

THEN
Block
```

## Content publishing

```text
WHEN
Action type equals Publish Article

THEN
Require approval from Content Editor
```

---

# 14. Non-functional requirements

## Security

The system must:

* encrypt secrets and continuation URLs;
* use TLS;
* hash exact action payloads;
* sign decision payloads;
* support expiring decisions;
* prevent replay;
* prevent duplicate decisions;
* verify workspace membership;
* enforce approver roles;
* prevent unauthorised self-approval;
* keep audit records immutable.

## Reliability

The system must:

* support idempotent action submission;
* retry webhook deliveries;
* survive service restarts;
* avoid losing approved decisions;
* process concurrent approval attempts safely;
* maintain delivery outbox records;
* expose failed deliveries to administrators.

## Performance

Target MVP performance:

* action submission response under 500 milliseconds when policy evaluation is immediate;
* inbox page load under 2 seconds for normal workspace sizes;
* decision delivery queued within 1 second;
* Slack notification queued within 2 seconds.

## Scalability

Initial target:

* 100 workspaces;
* 100,000 approval requests per month;
* up to 100 active policies per workspace;
* up to 50 users per workspace.

## Availability

Initial target:

* 99.5% monthly availability.

---

# 15. Security architecture

## Action hashing

The product must create a deterministic canonical representation of the action.

```text
original_action_hash = SHA-256(canonical_action_json)
```

For edited actions:

```text
approved_action_hash = SHA-256(canonical_approved_action_json)
```

## Decision signing

Signed data should include:

* workspace;
* request ID;
* decision ID;
* decision;
* original action hash;
* approved action hash;
* destination;
* issued time;
* expiration time.

## Replay protection

The source workflow should process each decision ID only once.

## Idempotency

Every request must include a stable idempotency key.

Example:

```text
n8n:execution-456:refund-1
```

Repeated submissions using the same key must return the existing request.

---

# 16. Suggested technical architecture

## Frontend

Possible stack:

* React;
* Next.js;
* TypeScript.

Pages:

* Login;
* Workspace setup;
* Inbox;
* Approval detail;
* Policies;
* Integrations;
* Team;
* Audit log;
* API keys.

## Backend

Recommended stack:

* Go;
* REST API;
* background workers;
* PostgreSQL;
* optional Redis later.

Core services:

* Authentication service;
* Workspace service;
* Approval request service;
* Policy evaluation service;
* Decision service;
* Delivery service;
* Notification service;
* Audit service;
* Integration service.

## Database

Suggested tables:

```text
users
workspaces
workspace_members
api_keys
agents
workflows
policies
policy_conditions
approver_groups
approver_group_members
approval_requests
approval_request_evidence
approval_decisions
approval_comments
continuation_targets
decision_deliveries
workflow_acknowledgements
notifications
audit_events
```

## Background processing

Use a durable outbox pattern for:

* webhook deliveries;
* Slack notifications;
* expiration jobs;
* escalation jobs.

PostgreSQL-backed jobs are sufficient for the MVP.

---

# 17. Suggested database entities

## Approval request

Important fields:

```text
id
workspace_id
action_type
title
description
original_action
original_action_hash
source_platform
source_workflow_id
source_execution_id
matched_policy_id
matched_policy_snapshot
assigned_approver_group_id
status
deadline_at
created_at
resolved_at
```

## Approval decision

Important fields:

```text
id
approval_request_id
decision
original_action_hash
approved_action
approved_action_hash
decided_by
comment
created_at
expires_at
```

## Continuation target

Important fields:

```text
id
approval_request_id
strategy
platform
external_execution_id
encrypted_configuration
created_at
```

## Decision delivery

Important fields:

```text
id
decision_id
status
attempt_count
next_attempt_at
last_attempt_at
last_response_code
last_error
delivered_at
acknowledged_at
```

---

# 18. MVP screens

## 18.1 Inbox

Displays:

* pending requests;
* assigned requests;
* deadlines;
* source platform;
* risk;
* action summary.

Primary action:

* Review.

## 18.2 Request detail

Displays:

* proposed action;
* evidence;
* original values;
* policy;
* deadline;
* approver;
* audit timeline.

Actions:

* Approve;
* Edit and approve;
* Reject.

## 18.3 Policies

Displays:

* policy name;
* conditions;
* result;
* approver;
* status;
* priority.

Actions:

* Create;
* Edit;
* Activate;
* Deactivate.

## 18.4 Integrations

Displays:

* API credentials;
* webhook secret;
* n8n template;
* LangGraph example;
* test integration status.

## 18.5 Audit log

Displays:

* timestamp;
* request;
* event;
* actor;
* platform;
* result.

---

# 19. Success metrics

## Activation

* percentage of workspaces that submit a first test request;
* percentage that connect a production workflow;
* time from signup to first approval request;
* number of policies created per workspace.

## Usage

* approval requests per workspace;
* weekly active approvers;
* number of connected workflows;
* percentage of requests resolved;
* average approval time;
* percentage approved, edited or rejected.

## Reliability

* decision-delivery success rate;
* callback retry rate;
* duplicate-request rate;
* permanently failed continuation rate;
* acknowledgement rate.

## Business validation

Strong MVP success signal:

* 3 to 5 active pilot customers;
* at least 10 connected production workflows;
* at least 500 real approval requests processed;
* at least one paying customer;
* at least one customer replacing custom approval logic;
* repeat weekly usage;
* more than one action type per customer.

The strongest validation signal is:

> A customer changes a production workflow so it cannot continue without receiving an Approval Inbox decision.

---

# 20. Risks and mitigations

## Risk: Existing platforms already offer approvals

Mitigation:

Focus on:

* cross-platform inbox;
* central policies;
* organisation-wide roles;
* audit history;
* multiple workflows and instances.

## Risk: Approval Inbox becomes only a feature

Mitigation:

Build value around:

* policy management;
* routing;
* evidence standardisation;
* auditability;
* workflow reliability;
* integration templates;
* exception analytics.

## Risk: Customers do not yet have enough agents

Mitigation:

Target automation agencies and companies already running production workflows rather than early AI experimenters.

## Risk: Approval fatigue

Mitigation:

Support:

* automatic allow;
* human approval;
* automatic block.

Do not require approval for every action.

## Risk: Workflow executes different parameters

Mitigation:

* action hashing;
* signed decision;
* return approved action;
* integration guidance;
* immutable original and approved payloads.

## Risk: Callback fails

Mitigation:

* durable delivery outbox;
* retry policy;
* manual retry;
* visible delivery state;
* polling fallback.

## Risk: LangGraph replay creates duplicate requests

Mitigation:

* stable idempotency keys;
* dedicated request-creation node;
* persistent request ID in graph state.

## Risk: Sensitive resume URL exposure

Mitigation:

* encrypt URLs;
* never expose them to frontend clients;
* remove them from logs;
* use signed callback payloads.

---

# 21. Development phases

## Phase 1: Core backend

Build:

* workspace model;
* authentication;
* API keys;
* approval-request API;
* idempotency;
* action hashing;
* policy evaluation;
* approval decisions;
* webhook delivery;
* audit events.

## Phase 2: Web interface

Build:

* inbox;
* request detail;
* approve;
* edit;
* reject;
* basic policy builder;
* team management.

## Phase 3: n8n integration

Build:

* importable template;
* Wait node flow;
* decision verification;
* refund example;
* documentation.

## Phase 4: Notifications

Build:

* Slack application;
* in-app notifications;
* approval links;
* deadline reminders.

## Phase 5: LangGraph integration

Build:

* Python helper;
* interrupt example;
* callback resume example;
* idempotency handling;
* durable checkpointer guidance.

## Phase 6: Pilot hardening

Build:

* callback retry dashboard;
* delivery failures;
* usage metrics;
* policy templates;
* onboarding improvements;
* audit export.

---

# 22. MVP acceptance criteria

The MVP is ready for pilot when:

1. A workspace administrator can register and invite an approver.
2. A developer can create and revoke an API key.
3. An n8n workflow can submit a proposed action.
4. The action can be evaluated against a policy.
5. An automatically allowed action receives an immediate response.
6. A blocked action receives an immediate block response.
7. An approval-required action appears in the inbox.
8. An assigned approver can approve the action.
9. An assigned approver can edit allowed parameters and approve.
10. An assigned approver can reject the action.
11. The original and approved action payloads remain immutable.
12. The product sends a signed decision to the n8n resume URL.
13. n8n resumes and routes according to the decision.
14. Failed callbacks are retried.
15. Duplicate submissions do not create duplicate requests.
16. Duplicate approval attempts do not create duplicate decisions.
17. Expired requests cannot be approved.
18. The audit log shows the complete request lifecycle.
19. Slack can notify the assigned approver.
20. A developer can integrate a generic application using a webhook or polling.

---

# 23. Initial positioning

## Headline

**One approval inbox for every AI workflow.**

## Supporting message

Define approval policies once, review risky actions from n8n, LangGraph and custom agents, and return the decision so the original workflow can continue.

## Developer message

**Stop rebuilding human approval logic into every automation.**

## Business message

**Deploy AI workflows without giving them unrestricted authority.**

---

# 24. Open questions

The following questions should be validated during customer interviews and pilots:

* Do customers want policies owned by developers or operations teams?
* Which action types generate the strongest willingness to pay?
* Is Slack approval sufficient for low-risk actions?
* Which evidence formats are most important?
* How often do approvers edit actions rather than approve or reject?
* Are multiple approvers needed in the first paid version?
* Do customers prefer webhook delivery or polling?
* How important is final workflow acknowledgement?
* Would automation agencies manage multiple client workspaces?
* What pricing metric is most understandable:

  * requests;
  * workflows;
  * agents;
  * workspaces?

---

# 25. Final MVP definition

The MVP is a centralised approval service where:

1. A workflow proposes an action.
2. Approval Inbox evaluates central policies.
3. Routine actions are allowed automatically.
4. Prohibited actions are blocked.
5. Exceptional actions enter a human inbox.
6. The approver reviews evidence.
7. The approver approves, edits or rejects.
8. Approval Inbox sends a signed decision.
9. The original workflow resumes.
10. The source workflow performs the action or stops.

The product does not replace workflow engines.

It provides the shared human-control layer they are currently missing.
