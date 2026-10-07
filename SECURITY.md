# Security Policy

## Reporting a vulnerability

Please do not open a public issue for a security problem. Use GitHub's private
[security advisory](https://github.com/cirxineiaviegxas/ContractFault/security/advisories/new) form, or contact a
maintainer directly if the form is unavailable to you.

Include what an attacker gains in one sentence, the version or commit you are running, the smallest
reproduction you can manage, and whether the input is attacker-controlled in your setup. You should
get an acknowledgement within a few working days and a disclosure timeline before anything is
published.

## Scope

ContractFault diffs two versions of a published contract and maps every change onto the
consumers it can reach. In scope: a crafted contract pair (duplicate consumer bindings, a
renamed field that survives in examples/) that produces a misleading blast radius, a change
that reaches a consumer the report does not list, a malformed contract that crashes the
analyzer instead of being rejected, and path handling in the report reader.
