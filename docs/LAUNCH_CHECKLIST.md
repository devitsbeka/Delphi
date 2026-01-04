# Delphi Launch Checklist

Use this checklist to ensure a successful production launch.

## Pre-Launch (1 Week Before)

### Infrastructure
- [ ] Production database provisioned (Supabase/Fly Postgres)
- [ ] Redis instance provisioned
- [ ] DNS configured for api.delphi.dev and app.delphi.dev
- [ ] SSL certificates verified
- [ ] CDN configured for static assets
- [ ] Backup strategy implemented and tested
- [ ] Disaster recovery plan documented

### Security
- [ ] All secrets rotated from development values
- [ ] JWT secret is 256-bit random
- [ ] API keys encrypted at rest
- [ ] Rate limiting configured
- [ ] CORS configured for production domains only
- [ ] Security headers enabled
- [ ] Vulnerability scan completed (Snyk/Dependabot)
- [ ] Penetration testing completed (if applicable)
- [ ] SOC 2 compliance verified (if applicable)

### Backend
- [ ] All migrations applied to production database
- [ ] Database indexes optimized
- [ ] Connection pooling configured
- [ ] Health check endpoints working
- [ ] Error handling comprehensive
- [ ] Logging configured for production
- [ ] Metrics exposed for monitoring

### Frontend
- [ ] Production build tested
- [ ] Environment variables configured
- [ ] Analytics integrated (if applicable)
- [ ] Error tracking configured (Sentry)
- [ ] Performance audit passed (Lighthouse 90+)
- [ ] Mobile responsiveness verified
- [ ] Browser compatibility tested

### Integrations
- [ ] GitHub OAuth app configured for production
- [ ] Stripe webhooks pointing to production
- [ ] Email service configured (SendGrid/Postmark)
- [ ] Slack/Discord webhooks configured

### Documentation
- [ ] API documentation complete
- [ ] User guide published
- [ ] Deployment runbook finalized
- [ ] Support procedures documented
- [ ] FAQ prepared

---

## Launch Day

### Deployment
- [ ] Final code review completed
- [ ] All tests passing
- [ ] Staging environment verified
- [ ] Database backup taken
- [ ] Team notified of launch window
- [ ] On-call engineer assigned

### Deployment Steps

```bash
# 1. Deploy backend
cd backend
fly deploy --app delphi-api

# 2. Verify backend health
curl https://api.delphi.dev/health

# 3. Deploy frontend
cd frontend
vercel --prod

# 4. Verify frontend loading
open https://app.delphi.dev

# 5. Run smoke tests
npm run test:smoke
```

### Smoke Tests
- [ ] Homepage loads
- [ ] Login works
- [ ] Registration works
- [ ] Dashboard displays data
- [ ] Agent creation works
- [ ] Agent execution works
- [ ] Repository connection works
- [ ] Knowledge base query works
- [ ] Settings page accessible
- [ ] Billing page accessible

### Monitoring
- [ ] Grafana dashboard accessible
- [ ] Alerts configured and tested
- [ ] Error tracking receiving events
- [ ] Logs streaming correctly

---

## Post-Launch (First 24 Hours)

### Monitoring Checklist
- [ ] Check error rates every hour
- [ ] Monitor response times
- [ ] Watch database connections
- [ ] Review user signups
- [ ] Check for failed payments
- [ ] Monitor agent execution success rate

### Key Metrics to Watch

| Metric | Threshold | Action |
|--------|-----------|--------|
| Error rate | > 1% | Investigate immediately |
| P99 latency | > 2s | Check database/cache |
| Signup failures | > 5% | Check auth flow |
| Agent failures | > 10% | Check provider APIs |
| Payment failures | > 5% | Check Stripe |

### Escalation Contacts

| Issue | Contact | Method |
|-------|---------|--------|
| Infrastructure | DevOps Lead | Slack #infrastructure |
| Application | Engineering Lead | Slack #engineering |
| Billing | Finance Lead | Slack #billing |
| Security | Security Lead | Phone |

---

## Week 1 Review

### Performance Review
- [ ] Analyze response time trends
- [ ] Review error patterns
- [ ] Check resource utilization
- [ ] Evaluate scaling needs

### User Feedback
- [ ] Review support tickets
- [ ] Analyze user behavior
- [ ] Collect feature requests
- [ ] Identify pain points

### Optimizations
- [ ] Address performance bottlenecks
- [ ] Fix reported bugs
- [ ] Implement quick wins
- [ ] Plan feature improvements

---

## Rollback Procedure

If critical issues occur:

```bash
# 1. Assess severity
# Critical: Data loss, security breach, complete outage
# High: Major feature broken, significant performance degradation
# Medium: Minor feature broken, edge cases failing

# 2. For critical issues, rollback immediately
fly releases rollback --app delphi-api

# 3. Notify team
# Post in #incidents channel with:
# - What happened
# - Impact
# - Actions taken
# - Next steps

# 4. Post-mortem within 48 hours
```

---

## Communication Templates

### Launch Announcement

```
🔮 Delphi is Live!

We're excited to announce the launch of Delphi, your AI Agent Command Center.

Start orchestrating dozens of AI agents across your repositories today:
- Multiple AI providers (OpenAI, Anthropic, Google, Ollama)
- GitHub integration with smart branch strategy
- Central knowledge base for context
- Beautiful Palantir-style dashboards

Get started: https://app.delphi.dev

Questions? Join our Discord: discord.gg/delphi
```

### Incident Communication

```
⚠️ Incident Notice

We're currently experiencing [ISSUE DESCRIPTION].

Impact: [AFFECTED SERVICES]
Status: [INVESTIGATING/IDENTIFIED/RESOLVED]
Next Update: [TIME]

We apologize for any inconvenience.

Status page: https://status.delphi.dev
```

### Resolution Communication

```
✅ Incident Resolved

The issue affecting [SERVICES] has been resolved.

Root Cause: [BRIEF DESCRIPTION]
Resolution: [WHAT WAS DONE]
Duration: [START - END TIME]

We've implemented [PREVENTIVE MEASURES] to prevent recurrence.

Thank you for your patience.
```

---

## Success Criteria

Launch is considered successful when:

### Day 1
- [ ] Zero critical bugs
- [ ] Error rate < 1%
- [ ] P99 latency < 2s
- [ ] 100+ signups

### Week 1
- [ ] 500+ signups
- [ ] 50+ paid conversions
- [ ] NPS > 30
- [ ] Support response < 4 hours

### Month 1
- [ ] 2000+ signups
- [ ] 200+ paid customers
- [ ] MRR > $10,000
- [ ] Uptime > 99.9%

---

## Post-Launch Improvements

### Priority Queue
1. Address all critical/high bugs
2. Implement most-requested features
3. Optimize performance bottlenecks
4. Improve onboarding flow
5. Add more AI providers

### Technical Debt
- [ ] Increase test coverage to 80%
- [ ] Refactor identified code smells
- [ ] Update dependencies
- [ ] Document undocumented APIs

---

<p align="center">
  Good luck with launch! 🚀🔮
</p>

