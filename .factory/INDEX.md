# Factory Configuration Index

Complete list of files and resources for DataRobot CLI Cloud Template setup.

## Setup Scripts

### `cloud-template-setup.sh`
**Full setup with complete validation**
- Task: `dev-init` → `lint` → `build` → `test`
- Duration: ~3-5 minutes
- Best for: Team reference, production-ready environments, pre-deployment checks
- When to use: First-time setup, CI/CD validation, release preparation

### `cloud-template-setup-minimal.sh`
**Fast setup (essentials only)**
- Task: `dev-init` → `build`
- Duration: ~1-2 minutes
- Best for: Quick iterations, development sessions
- When to use: Daily development, rapid prototyping

## Documentation

### `QUICK_START.md` ⭐ START HERE
**2-minute quick reference**
- Ready-to-copy setup scripts
- Step-by-step creation (Factory UI)
- Essential commands
- Quick troubleshooting
- 👉 **Best for:** Getting started fast

### `CLOUD_TEMPLATE_GUIDE.md`
**Comprehensive setup guide**
- What is a Cloud Template
- Detailed creation steps
- Setup script explanations
- Environment variables
- Best practices
- Customization guide
- Troubleshooting table
- 👉 **Best for:** Understanding the full picture

### `README.md`
**Factory configuration overview**
- Available scripts comparison table
- All Task commands reference
- Script customization examples
- Full troubleshooting guide
- Best practices
- 👉 **Best for:** Implementation details, customization

### `INDEX.md` (this file)
**Navigation and reference**
- File descriptions
- Usage recommendations
- Quick decision matrix
- 👉 **Best for:** Finding what you need

## Quick Decision Matrix

### "I want to..."

| Goal | Read This | Use This |
| ---- | --------- | -------- |
| Get started in 2 min    | QUICK_START.md          | cloud-template-setup-minimal.sh |
| Set up team environment | CLOUD_TEMPLATE_GUIDE.md | cloud-template-setup.sh         |
| Understand everything   | README.md               | Both scripts                    |
| Find a specific command | QUICK_START.md          | See "Common Commands"           |
| Customize the setup     | README.md               | Edit either script              |
| Troubleshoot issues     | CLOUD_TEMPLATE_GUIDE.md | See "Troubleshooting" section   |
| Share with teammates    | QUICK_START.md          | Copy the setup script           |

## File Structure

```
.factory/
├── INDEX.md (this file)
├── QUICK_START.md
├── CLOUD_TEMPLATE_GUIDE.md
├── README.md
├── cloud-template-setup.sh
└── cloud-template-setup-minimal.sh
```

## Setup Flow Diagram

```
Choose Setup Type
    ↓
┌─────────────────────────────────────────┐
│ Fast (1-2 min)    │    Full (3-5 min)   │
│ Minimal setup     │    With validation  │
│ dev-init+build    │    All checks+tests │
└─────────────────────────────────────────┘
    ↓
Copy Script to Factory UI
    ↓
Create Cloud Template
    ↓
Wait for "Ready" Status (~1-5 min)
    ↓
Launch from Factory Session
    ↓
Start Coding!
```

## External Resources

- **Factory Cloud Templates Docs:** https://docs.factory.ai/web/machine-connection/cloud-templates
- **DataRobot CLI README:** See project root `README.md`
- **AGENTS.md:** Project coding guidelines and conventions
- **Taskfile.yaml:** Complete task definitions in project root
- **Task Runner Docs:** https://taskfile.dev/

## Common Tasks & Commands

### Create a Template
1. Factory Settings → Cloud Templates
2. Click "Create Template"
3. Copy script from `QUICK_START.md`
4. Submit

### Use a Template
1. Start Factory session
2. Machine Connection → Remote
3. Select your template
4. Connect and code

### Run Tasks
```bash
task run              # Execute CLI
task build            # Build binary
task test             # Run tests
task lint             # Code quality checks
task dev-init         # Initialize environment
```

## Troubleshooting Quick Links

| Issue | Solution |
| ----- | -------- |
| Setup fails       | → See CLOUD_TEMPLATE_GUIDE.md "Troubleshooting"   |
| Task not found    | → Check PATH or restart terminal                  |
| Need help         | → Read README.md "Troubleshooting" table          |
| Want faster setup | → Use cloud-template-setup-minimal.sh             |
| Want validation   | → Use cloud-template-setup.sh                     |

## Tips

1. **Start with QUICK_START.md** - Most users can follow this in 2 minutes
2. **Use minimal setup first** - Get familiar with Cloud Templates quickly
3. **Try full setup next** - Understand the validation workflow
4. **Customize later** - Once comfortable, tweak the scripts for your workflow
5. **Share with team** - Copy the template URL from Factory settings

## Feedback & Customization

These files are part of your repository (in `.factory/`), so you can:
- Edit scripts to match your team's workflow
- Update docs as processes change
- Version control all changes with git
- Share templates with your entire team

## Questions?

1. Check QUICK_START.md first (fastest)
2. Read CLOUD_TEMPLATE_GUIDE.md (comprehensive)
3. Review README.md (detailed reference)
4. Visit Factory docs: https://docs.factory.ai/web/machine-connection/cloud-templates