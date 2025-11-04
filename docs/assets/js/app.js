// App State
const appState = {
    sidebarOpen: window.innerWidth >= 1024, // Open by default on desktop
    expandedSections: new Set(['architecture', 'database', 'deployment']), // Default expanded
    currentPath: null
};

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    initializeApp();
});

function initializeApp() {
    // Initialize sidebar state (open on desktop)
    const sidebar = document.getElementById('sidebar');
    if (appState.sidebarOpen) {
        sidebar.classList.add('open');
    }
    
    // Setup sidebar toggle
    const sidebarToggle = document.getElementById('sidebarToggle');
    
    sidebarToggle.addEventListener('click', () => {
        toggleSidebar();
    });

    // Setup navigation section headers (expand/collapse)
    setupNavigationSections();

    // Setup navigation links
    setupNavigationLinks();

    // Setup routing
    setupRouting();

    // Handle initial route
    handleRoute();
    
    // Setup overlay for mobile
    setupOverlay();
    
    // Handle window resize
    window.addEventListener('resize', () => {
        const shouldBeOpen = window.innerWidth >= 1024;
        if (shouldBeOpen !== appState.sidebarOpen) {
            appState.sidebarOpen = shouldBeOpen;
            sidebar.classList.toggle('open', shouldBeOpen);
        }
    });
}

function toggleSidebar() {
    const sidebar = document.getElementById('sidebar');
    const overlay = document.querySelector('.sidebar-overlay');
    
    appState.sidebarOpen = !appState.sidebarOpen;
    sidebar.classList.toggle('open', appState.sidebarOpen);
    
    if (overlay) {
        overlay.classList.toggle('active', appState.sidebarOpen);
    }

    // Close on mobile when clicking outside
    if (window.innerWidth <= 1023 && appState.sidebarOpen) {
        overlay.addEventListener('click', () => {
            toggleSidebar();
        }, { once: true });
    }
}

function setupOverlay() {
    const overlay = document.createElement('div');
    overlay.className = 'sidebar-overlay';
    document.body.appendChild(overlay);
}

function setupNavigationSections() {
    const sectionHeaders = document.querySelectorAll('.nav-section-header');
    
    sectionHeaders.forEach(header => {
        const section = header.getAttribute('data-section');
        const subsection = document.querySelector(`.nav-subsection[data-subsection="${section}"]`);
        
        // Set initial expanded state
        if (appState.expandedSections.has(section)) {
            header.classList.add('expanded');
            subsection.classList.add('expanded');
        }

        header.addEventListener('click', (e) => {
            e.preventDefault();
            toggleSection(section);
        });
    });
}

function toggleSection(section) {
    const header = document.querySelector(`.nav-section-header[data-section="${section}"]`);
    const subsection = document.querySelector(`.nav-subsection[data-subsection="${section}"]`);
    
    const isExpanded = header.classList.contains('expanded');
    
    if (isExpanded) {
        header.classList.remove('expanded');
        subsection.classList.remove('expanded');
        appState.expandedSections.delete(section);
    } else {
        header.classList.add('expanded');
        subsection.classList.add('expanded');
        appState.expandedSections.add(section);
    }
}

function setupNavigationLinks() {
    const navLinks = document.querySelectorAll('.nav-link');
    
    navLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const path = link.getAttribute('data-path');
            navigateTo(path);
        });
    });
}

function setupRouting() {
    // Handle hash changes
    window.addEventListener('hashchange', () => {
        handleRoute();
    });

    // Handle browser back/forward
    window.addEventListener('popstate', () => {
        handleRoute();
    });
}

function handleRoute() {
    const hash = window.location.hash || '#/';
    const path = hash.substring(2); // Remove '#/'
    
    if (!path) {
        navigateTo('index.md');
        return;
    }

    // Map route to markdown file
    const routeMap = {
        '': 'index.md',
        'architecture/getting-started': 'architecture/getting-started.md',
        'architecture/overview': 'architecture/overview.md',
        'architecture/api-layer-details': 'architecture/api-layer-details.md',
        'architecture/core-api': 'architecture/core-api.md',
        'architecture/data-type-detection-patterns': 'architecture/data-type-detection-patterns.md',
        'architecture/database-schemas': 'architecture/database-schemas.md',
        'architecture/diagrams': 'architecture/diagrams.md',
        'database/BACKUP_MONGODB': 'database/BACKUP_MONGODB.md',
        'database/BACKUP_NEO4J': 'database/BACKUP_NEO4J.md',
        'database/BACKUP_POSTGRES': 'database/BACKUP_POSTGRES.md',
        'deployment/BACKUP_INTEGRATION': 'deployment/BACKUP_INTEGRATION.md',
        'how_it_works': 'how_it_works.md',
        'datatype': 'datatype.md',
        'storage': 'storage.md',
        'release_life_cycle': 'release_life_cycle.md',
        'limitations': 'limitations.md'
    };

    const filePath = routeMap[path] || routeMap[''];
    navigateTo(filePath);
}

function navigateTo(filePath) {
    appState.currentPath = filePath;
    
    // Update active link
    updateActiveLink(filePath);
    
    // Auto-expand parent section if viewing a child page
    autoExpandParentSection(filePath);
    
    // Load and render markdown
    loadMarkdown(filePath);
}

function updateActiveLink(filePath) {
    const navLinks = document.querySelectorAll('.nav-link');
    navLinks.forEach(link => {
        const linkPath = link.getAttribute('data-path');
        if (linkPath === filePath) {
            link.classList.add('active');
        } else {
            link.classList.remove('active');
        }
    });
}

function autoExpandParentSection(filePath) {
    // Auto-expand architecture section
    if (filePath.startsWith('architecture/')) {
        if (!appState.expandedSections.has('architecture')) {
            toggleSection('architecture');
        }
    }
    
    // Auto-expand database section
    if (filePath.startsWith('database/')) {
        if (!appState.expandedSections.has('database')) {
            toggleSection('database');
        }
    }
    
    // Auto-expand deployment section
    if (filePath.startsWith('deployment/')) {
        if (!appState.expandedSections.has('deployment')) {
            toggleSection('deployment');
        }
    }
}

async function loadMarkdown(filePath) {
    const contentArea = document.getElementById('markdownContent');
    contentArea.innerHTML = '<div class="loading">Loading documentation...</div>';
    
    try {
        const response = await fetch(filePath);
        if (!response.ok) {
            throw new Error(`Failed to load ${filePath}`);
        }
        
        const markdown = await response.text();
        renderMarkdown(markdown, filePath);
    } catch (error) {
        console.error('Error loading markdown:', error);
        contentArea.innerHTML = `
            <div style="padding: 40px; text-align: center;">
                <h2 style="color: var(--orange);">Error Loading Document</h2>
                <p>${error.message}</p>
                <p><a href="#/">Return to Home</a></p>
            </div>
        `;
    }
}

function renderMarkdown(markdown, filePath) {
    const contentArea = document.getElementById('markdownContent');
    
    // Configure marked options
    marked.setOptions({
        breaks: true,
        gfm: true,
        highlight: function(code, lang) {
            if (lang && hljs.getLanguage(lang)) {
                try {
                    return hljs.highlight(code, { language: lang }).value;
                } catch (err) {
                    console.error('Error highlighting code:', err);
                }
            }
            return hljs.highlightAuto(code).value;
        }
    });
    
    // Convert markdown to HTML
    let html = marked.parse(markdown);
    
    // Process relative links in markdown
    html = processMarkdownLinks(html, filePath);
    
    // Remove frontmatter if present
    html = html.replace(/^---[\s\S]*?---\n/, '');
    
    contentArea.innerHTML = html;
    
    // Highlight code blocks
    hljs.highlightAll();
    
    // Scroll to top
    window.scrollTo(0, 0);
}

function processMarkdownLinks(html, currentFilePath) {
    // Create a temporary DOM element to parse HTML
    const tempDiv = document.createElement('div');
    tempDiv.innerHTML = html;
    
    const links = tempDiv.querySelectorAll('a');
    links.forEach(link => {
        const href = link.getAttribute('href');
        
        if (!href) return;
        
        // Skip external links
        if (href.startsWith('http://') || href.startsWith('https://') || href.startsWith('mailto:')) {
            return;
        }
        
        // Skip hash-only links
        if (href.startsWith('#')) {
            return;
        }
        
        // Convert relative markdown links to hash routes
        if (href.endsWith('.md') || (!href.includes('://') && !href.startsWith('/'))) {
            let route = href;
            
            // Handle relative paths
            if (href.startsWith('./')) {
                const currentDir = currentFilePath.substring(0, currentFilePath.lastIndexOf('/') + 1);
                route = currentDir + href.substring(2);
            } else if (href.startsWith('../')) {
                // Handle parent directory references
                let currentDir = currentFilePath.substring(0, currentFilePath.lastIndexOf('/'));
                const parts = href.split('../');
                for (let i = 0; i < parts.length - 1; i++) {
                    currentDir = currentDir.substring(0, currentDir.lastIndexOf('/'));
                }
                route = currentDir + '/' + parts[parts.length - 1];
            } else if (!href.startsWith('/')) {
                // Relative to current directory
                if (currentFilePath.includes('/')) {
                    const currentDir = currentFilePath.substring(0, currentFilePath.lastIndexOf('/') + 1);
                    route = currentDir + href;
                } else {
                    route = href;
                }
            } else {
                route = href.substring(1);
            }
            
            // Ensure .md extension
            if (!route.endsWith('.md') && !route.includes('.')) {
                route += '.md';
            }
            
            // Normalize the route
            route = route.replace(/\/+/g, '/').replace(/^\.\//, '');
            
            // Map to hash route
            const routeToHash = (filePath) => {
                const cleanPath = filePath.replace(/\.md$/, '');
                const routeMap = {
                    'index': '',
                    'architecture/getting-started': 'architecture/getting-started',
                    'architecture/overview': 'architecture/overview',
                    'architecture/api-layer-details': 'architecture/api-layer-details',
                    'architecture/core-api': 'architecture/core-api',
                    'architecture/data-type-detection-patterns': 'architecture/data-type-detection-patterns',
                    'architecture/database-schemas': 'architecture/database-schemas',
                    'architecture/diagrams': 'architecture/diagrams',
                    'database/BACKUP_MONGODB': 'database/BACKUP_MONGODB',
                    'database/BACKUP_NEO4J': 'database/BACKUP_NEO4J',
                    'database/BACKUP_POSTGRES': 'database/BACKUP_POSTGRES',
                    'deployment/BACKUP_INTEGRATION': 'deployment/BACKUP_INTEGRATION',
                    'how_it_works': 'how_it_works',
                    'datatype': 'datatype',
                    'storage': 'storage',
                    'release_life_cycle': 'release_life_cycle',
                    'limitations': 'limitations'
                };
                return routeMap[cleanPath] !== undefined ? routeMap[cleanPath] : cleanPath;
            };
            
            const hashRoute = routeToHash(route);
            link.setAttribute('href', `#/${hashRoute}`);
            
            // Remove existing listeners and add new one
            const newLink = link.cloneNode(true);
            link.parentNode.replaceChild(newLink, link);
            newLink.addEventListener('click', (e) => {
                e.preventDefault();
                navigateTo(route);
            });
        }
    });
    
    return tempDiv.innerHTML;
}

// Export functions for potential external use
window.docsApp = {
    navigateTo,
    toggleSidebar
};

