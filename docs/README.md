# Alex Project Website

This directory contains the official website for the Alex AI Coding Assistant project.

## 🌟 Features

### Ultra-Thin Design
- Minimalist, clean interface with subtle borders and thin typography
- Modern grid layouts with optimal spacing
- Focus on content hierarchy and readability

### Sci-Fi Tech Aesthetic  
- Dark theme with neon accent colors (#00d4ff)
- Gradient backgrounds and glowing effects
- Futuristic typography (JetBrains Mono + Inter)
- Animated background particles

### Interactive Elements
- Smooth animations and transitions
- Typing effect in terminal demo
- Hover effects on feature cards
- Parallax background scrolling
- Responsive design for all devices

## 🚀 Quick Start

### Local Development
```bash
# Navigate to docs directory
cd docs/

# Serve locally (Python 3)
python -m http.server 8000

# Or with Node.js
npx serve .

# Open in browser
open http://localhost:8000
```

### Deploy to GitHub Pages
1. Push the `docs/` folder to your repository
2. Go to Settings > Pages
3. Set Source to "Deploy from a branch"
4. Select "main" branch and "/docs" folder
5. Your site will be available at `https://yourusername.github.io/Alex-Code`

## 📁 File Structure

```
docs/
├── index.html          # Main landing page
├── README.md           # This file
└── assets/             # Static assets (if needed)
    ├── images/
    └── icons/
```

## 🎨 Design System

### Colors
- Primary Background: `#0a0a0a`
- Secondary Background: `#111111`
- Accent Color: `#00d4ff` (Neon Blue)
- Success Color: `#00ff88` (Neon Green)
- Text Primary: `#ffffff`
- Text Secondary: `#a0a0a0`

### Typography
- Headlines: `Inter` (300-700 weights)
- Code/Terminal: `JetBrains Mono`
- Body Text: `Inter` (400-500 weights)

### Animations
- Fade-in on scroll
- Typing effect for terminal
- Background particle movement
- Smooth hover transitions
- Parallax scrolling

## 🛠 Customization

### Adding New Sections
1. Add HTML structure to `index.html`
2. Style with CSS following the existing design system
3. Add JavaScript interactions if needed

### Modifying Colors
Update CSS custom properties in the `:root` selector:
```css
:root {
    --accent-color: #your-color;
    --accent-glow: #your-color33;
}
```

### Performance Optimization
- All assets are optimized for fast loading
- CSS animations use `transform` for better performance
- Minimal JavaScript for interactions
- Web fonts loaded asynchronously

## 📱 Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## 🔧 Technical Notes

- Uses CSS Grid and Flexbox for layouts
- CSS Custom Properties for theming
- Intersection Observer for scroll animations
- No external dependencies except Google Fonts
- Fully responsive design

## 📈 Analytics (Optional)

To add analytics, insert your tracking code before the closing `</head>` tag:

```html
<!-- Google Analytics -->
<script async src="https://www.googletagmanager.com/gtag/js?id=GA_TRACKING_ID"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());
  gtag('config', 'GA_TRACKING_ID');
</script>
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes to the website
4. Test on multiple devices/browsers
5. Submit a pull request

## 📄 License

This website follows the same license as the Alex project (MIT License).

---

For more information about the Alex AI Coding Assistant, visit the [main repository](https://github.com/cklxx/Alex-Code).