# 🔒 Security Policy

Thank you for helping keep AI Village and its users safe!

## 📦 Supported Versions

AI Village is currently under active development. Until version 1.0.0 is released, security fixes will primarily target the latest development version.

| Version | Supported |
|---|---|
| Latest stable release | ✅ |
| Older releases | ❌ |
| Development and pre-release builds | Best effort |

Users are encouraged to update to the latest stable release.

## 🚨 Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub Issues, pull requests, or discussions.**

Instead, use GitHub's **[Private Vulnerability Reporting](https://github.com/FunisWokodeneer/AI-Village/security/advisories/new)** feature.

When submitting a report, please include:

- A description of the vulnerability.
- The affected version, component, or feature.
- The potential security impact.
- Steps to reproduce the issue, if possible.
- Relevant logs, diagnostic information, or suspected code locations.
- Suggested fixes or mitigations, if known.

Please do not include passwords, API keys, access tokens, or unrelated private information.

### ⏱️ Response Timeline

We aim to:

- Acknowledge reports within 7 days.
- Provide an initial assessment within 14 days.
- Keep reporters informed of significant progress.
- Work toward a fix or mitigation as soon as reasonably possible.

These are target timelines, not guaranteed deadlines.

## 🔍 Security Monitoring and Logging

AI Village is designed to maintain comprehensive application-wide logs to support bug reports, debugging, security investigations, and accountability.

Where applicable, logging should cover:

- User actions and configuration changes.
- AI agent actions, decisions, assignments, and delegation.
- AI model requests, responses, and provider errors.
- Tool calls, command execution, and results.
- File operations and workspace access.
- Permission checks, approvals, and denied actions.
- Application requests, errors, exceptions, and crashes.
- Security warnings, vulnerability detections, and recovery events.

Logs should contain timestamps, relevant identifiers, operation details, outcomes, and diagnostic information sufficient to investigate incidents.

Logs must be searchable, filterable, and exportable for bug reports and security investigations.

### 🔐 Log Protection and Privacy

- Logs must be protected against unauthorized access, modification, and deletion.
- Agents must not be able to silently alter or erase their own audit history.
- Credentials, passwords, API keys, and access tokens must never intentionally be stored in plaintext logs.
- Sensitive information must be redacted where appropriate.
- Private file contents should not be unnecessarily recorded.
- Log retention, rotation, and archival must follow documented policies.

**Logging must continue during security lockdown.** Stopping agent activity must never mean losing the diagnostic information needed to investigate an incident.

## 🛑 Emergency Security Response

AI Village is intended to include an application-level security monitoring and emergency shutdown mechanism.

When a credible potential vulnerability is detected by an AI agent, automated scanner, or application monitoring component:

1. AI Village must immediately enter a global security lockdown.
2. All active AI agents must stop, including agents in the middle of tasks.
3. New assignments, delegation, and potentially unsafe tool or command execution must stop.
4. Relevant task state, diagnostic information, and logs must be preserved.
5. The person who launched the application and is responsible for maintaining it must be notified.
6. A security incident record must be created.

The incident record should include the detection source, timestamp, affected components, suspected code locations, relevant evidence, severity assessment, and interrupted tasks.

### 📡 Automatic Reporting

If automatic security reporting is enabled, AI Village should prepare and submit a private GitHub vulnerability report through an authorized mechanism, where supported.

If automatic reporting is unavailable or fails, the incident must remain available for manual submission.

The maintainer must be instructed to report the vulnerability privately through GitHub immediately.

Automatic reporting does not replace maintainer notification.

### 🛠️ Recovery

AI Village must remain in security lockdown until:

1. The finding has been investigated.
2. A fix or adequate mitigation has been implemented.
3. The fix or mitigation has been independently verified.
4. The required private GitHub report has been submitted or linked to the incident.
5. The maintainer has reviewed the incident and explicitly authorized recovery.

If a finding is determined to be a false positive, the maintainer must document and review that decision before unlocking the application.

AI agents must never independently clear a security lockdown.

Restarting the application must not automatically clear an unresolved incident. Interrupted tasks must not automatically resume after recovery.

**Logging and diagnostics must remain operational throughout the entire lockdown and recovery process.**

## 📋 Security Incident Handling

Security incidents should be tracked with clear statuses:

| Status | Description |
|---|---|
| 🔴 Detected | A potential vulnerability was identified. |
| 🔍 Investigating | The finding is being reviewed. |
| ⚠️ Confirmed | A vulnerability has been verified. |
| 🛠️ Mitigated | A fix or mitigation has been implemented. |
| 📡 Reported | A private GitHub report has been submitted or linked. |
| 🟢 Resolved | Recovery has been authorized by the maintainer. |

Incident records and relevant logs should remain available after recovery.

## 📌 Additional Information

This policy applies to the AI Village application, its official source code, and security-related functionality maintained in this repository.

Third-party services, AI models, libraries, and infrastructure may have separate security reporting procedures. Vulnerabilities affecting those systems should also be reported to their respective maintainers.

---

**🔒 Security first. Transparency always.**
