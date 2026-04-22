# Security Policy

## Supported Versions

We take security seriously. The following versions of kmgr are currently supported with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in kmgr, please report it to us as follows:

1. **DO NOT** create a public GitHub issue
2. Email security@kmgr.dev (create this email alias or use your personal email for now)
3. Include detailed information about the vulnerability:
   - Description of the issue
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

## Security Considerations

kmgr handles sensitive Kubernetes configuration files. Here are the security measures we implement:

- **File Permissions**: All kubeconfig files are created with 0600 permissions
- **No External Dependencies**: Standalone binary with no runtime dependencies
- **Input Validation**: Strict validation of kubeconfig files and naming conventions
- **Backup System**: Automatic backups before any modifications
- **Quarantine**: Invalid files are automatically moved to quarantine directory

## Responsible Disclosure

We kindly ask that you:

- Give us reasonable time to fix the issue before public disclosure
- Avoid accessing or modifying user data
- Respect the confidentiality of the report

We will acknowledge receipt of your report within 48 hours and provide a more detailed response within 7 days indicating our next steps.