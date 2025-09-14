---
layout: default
title: "Project Management"
description: "Development status, roadmaps, and project planning"
collection: project-management
permalink: /project-management/
---

# Project Management

This section contains project planning documents, development status updates, and roadmap information for the Credence project.

## Documents

{% for doc in site.project-management %}
- **[{{ doc.title }}]({{ doc.url | relative_url }})** - {{ doc.description }}
{% endfor %}

## Current Status

The Credence project is actively under development with core services implemented and deployment tooling in progress. See individual component status pages for detailed information.

## Project Timeline

- **Phase 1**: Core Infrastructure ✅
- **Phase 2**: Network Services 🔄
- **Phase 3**: Client Applications 🔄  
- **Phase 4**: Production Deployment 🔮