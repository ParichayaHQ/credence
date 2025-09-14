# Credence Documentation

This directory contains the documentation for the Credence decentralized trust network, configured for GitHub Pages with Jekyll.

## 📚 Documentation Structure

- **[/guides/](guides/)** - Step-by-step deployment and usage guides
- **[/designs/](designs/)** - Architecture and technical design documents  
- **[/project-management/](project-management/)** - Project status and planning

## 🚀 GitHub Pages Setup

This documentation is set up to be automatically published to GitHub Pages using Jekyll. The site will be available at:
`https://parichayahq.github.io/credence`

### Configuration Files

- `_config.yml` - Jekyll configuration for GitHub Pages
- `Gemfile` - Ruby gem dependencies
- `_layouts/` - Jekyll layout templates
- `_includes/` - Reusable Jekyll components
- `assets/css/` - Custom styling

### Collections

The docs are organized into Jekyll collections:

- `_guides/` - Deployment and integration guides
- `_designs/` - Architecture documentation
- `_project-management/` - Project planning and status

## 🛠️ Local Development

To run the documentation site locally:

1. Install Ruby and Bundler
2. Install dependencies:
   ```bash
   cd docs/
   bundle install
   ```

3. Serve the site:
   ```bash
   bundle exec jekyll serve
   ```

4. Visit `http://localhost:4000/credence` to view the site

## 📝 Adding Content

### New Guide

1. Create a new file in `_guides/`
2. Add Jekyll front matter:
   ```yaml
   ---
   layout: guide
   title: "Your Guide Title"
   description: "Brief description"
   status: beta  # draft, beta, or stable
   collection: guides
   ---
   ```
3. Write your content in Markdown

### New Design Document

1. Create a new file in `_designs/`
2. Add Jekyll front matter:
   ```yaml
   ---
   layout: design
   title: "Document Title"
   description: "Brief description"
   collection: designs
   ---
   ```
3. Write your content in Markdown

## 🎨 Styling

Custom CSS is in `assets/css/custom.css` and includes:
- Responsive grid system for homepage cards
- Navigation styling
- Guide and design document formatting
- Status banners and badges
- Code syntax highlighting

## 🔧 Features

- **Responsive Design** - Mobile-friendly layouts
- **Navigation** - Automated sidebar navigation
- **Search** - GitHub Pages built-in search
- **SEO** - Meta tags and structured data
- **Collections** - Organized content with Jekyll collections
- **Breadcrumbs** - Navigation context for guides and designs
- **Edit Links** - Direct links to edit pages on GitHub
- **Status Badges** - Development status indicators

## 📊 Analytics

GitHub Pages provides basic analytics. For more detailed analytics, you can integrate Google Analytics by adding the tracking ID to `_config.yml`.

## 🚀 Deployment

The site is automatically deployed when changes are pushed to the `main` branch. GitHub Actions will:

1. Build the Jekyll site
2. Deploy to GitHub Pages
3. Make it available at the configured URL

No additional deployment steps are required!