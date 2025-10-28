package main

import (
	"fmt"
	"html"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/saintparish4/pantera/database"
	"github.com/saintparish4/pantera/routes"
)

func getAsciiArtOnly() string {
	return `
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++*+*+++++++++++++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++++++++++++          +++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++            ++++++++++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++++++++++*            ++++++++++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++++++++++              +++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++                ++++++++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++++++++        ++        +++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++        ++++        ++++++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++++++        +++++         +++++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++++++        ++++++        +++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++        ++++++++        ++++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++++        ++++++++++        +++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++        +++++++++++         ++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++        +++++++++++++         +++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++        ++++++++++++++        +++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++        ++++++++++++++++        ++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++        +++++++++++++++++         +++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++        +++++++++++++++++++         ++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++        ++++++++++++++++++++        ++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++        ++++++++++++++++++++++        +++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++        ++++++++++++++++++++++++        ++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++        +++++++++++++++++++++++++         +++++++++++++++++++++++++++++
++++++++++++++++++++++++++++        +++++++++++++++++++++++++++         ++++++++++++++++++++++++++++
+++++++++++++++++++++++++++*        ++++++++++++++++++++++++++++        *+++++++++++++++++++++++++++
+++++++++++++++++++++++++++        ++++++++++++++++++++++++++++++        +++++++++++++++++++++++++++
++++++++++++++++++++++++++        +++++++++++++++++++++++++++++++         ++++++++++++++++++++++++++
+++++++++++++++++++++++++        +++++++++++++++++++++++++++++++++         +++++++++++++++++++++++++
++++++++++++++++++++++++         ++++++++++++++++++++++++++++++++++         ++++++++++++++++++++++++
++++++++++++++++++++++++        ++++++++++++++++++++++++++++++++++++        ++++++++++++++++++++++++
+++++++++++++++++++++++        +++++++++++++++++++++++++++++++++++++         +++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
`
}

func getManifestoHTML() string {
	return `
  <p>In the wild, survival belongs to those who can read the terrain 
  before it shifts.</p>

  <p>Markets are no different.</p>

  <p>Every fluctuation. Every surge. Every unseen variable — hidden in the noise.<br>
  Most react. Few anticipate.</p>

  <p><strong>Pantera doesn't predict.<br>
  It hunts.</strong></p>

  <p>Forged in code and driven by instinct, Pantera transforms raw data into 
  precision decisions — in milliseconds.</p>

  <p>A dynamic pricing engine built not just to calculate, but to sense.</p>

  <p>Behind every transaction, a mind that adapts.<br>
  Behind every number, an algorithm that learns.</p>

  <p><strong>Fast. Silent. Relentless.</strong></p>

  <p><em>This isn't automation.<br>
  It's evolution.</em></p>

  <p style="font-size: 18px; margin-top: 20px;"><strong>Pantera — where instinct becomes intelligence.</strong></p>
`
}

func getBannerText() string {
	return `
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++*+*+++++++++++++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++++++++++++          +++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++            ++++++++++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++++++++++*            ++++++++++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++++++++++              +++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++                ++++++++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++++++++        ++        +++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++        ++++        ++++++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++++++        +++++         +++++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++++++        ++++++        +++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++        ++++++++        ++++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++++        ++++++++++        +++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++        +++++++++++         ++++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++        +++++++++++++         +++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++++        ++++++++++++++        +++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++        ++++++++++++++++        ++++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++++        +++++++++++++++++         +++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++        +++++++++++++++++++         ++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++        ++++++++++++++++++++        ++++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++++        ++++++++++++++++++++++        +++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++        ++++++++++++++++++++++++        ++++++++++++++++++++++++++++++
+++++++++++++++++++++++++++++        +++++++++++++++++++++++++         +++++++++++++++++++++++++++++
++++++++++++++++++++++++++++        +++++++++++++++++++++++++++         ++++++++++++++++++++++++++++
+++++++++++++++++++++++++++*        ++++++++++++++++++++++++++++        *+++++++++++++++++++++++++++
+++++++++++++++++++++++++++        ++++++++++++++++++++++++++++++        +++++++++++++++++++++++++++
++++++++++++++++++++++++++        +++++++++++++++++++++++++++++++         ++++++++++++++++++++++++++
+++++++++++++++++++++++++        +++++++++++++++++++++++++++++++++         +++++++++++++++++++++++++
++++++++++++++++++++++++         ++++++++++++++++++++++++++++++++++         ++++++++++++++++++++++++
++++++++++++++++++++++++        ++++++++++++++++++++++++++++++++++++        ++++++++++++++++++++++++
+++++++++++++++++++++++        +++++++++++++++++++++++++++++++++++++         +++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

                           P A N T E R A
                    Dynamic Pricing Engine v1.0.0

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  In the wild, survival belongs to those who can read the terrain 
  before it shifts.

  Markets are no different.

  Every fluctuation. Every surge. Every unseen variable — hidden in the noise.
  Most react. Few anticipate.

  Pantera doesn't predict.
  It hunts.

  Forged in code and driven by instinct, Pantera transforms raw data into 
  precision decisions — in milliseconds.

  A dynamic pricing engine built not just to calculate, but to sense.

  Behind every transaction, a mind that adapts.
  Behind every number, an algorithm that learns.

  Fast. Silent. Relentless.

  This isn't automation.
  It's evolution.

  Pantera — where instinct becomes intelligence.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`
}

func printBanner() {
	fmt.Println(getBannerText())
}

func main() {
	// Print startup banner
	printBanner()
	// Load environment variables from parent directory
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, trying current directory...")
		godotenv.Load() // Try current directory as fallback
	}

	// Initialize database
	database.Connect()
	database.Migrate()

	// Setup Gin router
	router := gin.Default()

	// CORS middleware for Frontend
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Root endpoint - Display banner to users
	router.GET("/", func(c *gin.Context) {
		// Check if client accepts HTML (browsers typically send text/html in Accept header)
		acceptHeader := c.GetHeader("Accept")
		wantHTML := strings.Contains(acceptHeader, "text/html") || c.Query("format") == "html"

		if wantHTML {
			htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <title>Pantera - Dynamic Pricing Engine</title>
    <meta charset="UTF-8">
    <style>
        body {
            background: #0a0a0a;
            color: #00ff00;
            font-family: 'Courier New', monospace;
            padding: 20px;
            line-height: 1.6;
            overflow-x: auto;
        }
        pre.ascii-art {
            white-space: pre;
            font-size: 8px;
            color: #ff0000;
            margin: 0 auto;
            line-height: 1.2;
            text-align: center;
            display: block;
            background: #0a0a0a;
            position: relative;
            z-index: 1;
        }
        .title {
            font-size: 32px;
            font-weight: bold;
            color: #ff0000;
            text-align: center;
            letter-spacing: 8px;
            margin: 20px 0 10px 0;
            background: #0a0a0a;
            position: relative;
            z-index: 1;
        }
        .subtitle {
            font-size: 20px;
            color: #ffffff;
            text-align: center;
            margin-bottom: 30px;
            background: #0a0a0a;
            position: relative;
            z-index: 1;
        }
        .manifesto {
            font-size: 24px;
            color: #cccccc;
            line-height: 1.8;
            max-width: 900px;
            margin: 0 auto 30px auto;
            text-align: left;
            position: relative;
            background: #0a0a0a;
            z-index: 1;
        }
        @keyframes glitch {
            0%, 100% { transform: translate(0); }
            20% { transform: translate(-2px, 2px); }
            40% { transform: translate(-2px, -2px); }
            60% { transform: translate(2px, 2px); }
            80% { transform: translate(2px, -2px); }
        }
        @keyframes glitch-text {
            0% { text-shadow: 0 0 0 #ff0000; }
            20% { text-shadow: -2px 0 0 #ff0000, 2px 0 0 #00ff00; }
            40% { text-shadow: 2px 0 0 #ff0000, -2px 0 0 #00ff00; }
            60% { text-shadow: 0 0 0 #ff0000; }
            80% { text-shadow: -2px 0 0 #00ff00, 2px 0 0 #ff0000; }
            100% { text-shadow: 0 0 0 #ff0000; }
        }
        .manifesto strong {
            animation: glitch-text 3s infinite;
            color: #ff0000;
        }
        .manifesto em {
            color: #00ff00;
        }
        .typing-cursor {
            display: inline-block;
            width: 3px;
            height: 20px;
            background: #00ff00;
            animation: blink 0.7s infinite;
            margin-left: 2px;
        }
        @keyframes blink {
            0%, 49% { opacity: 1; }
            50%, 100% { opacity: 0; }
        }
        .separator {
            color: #ff0000;
            text-align: center;
            font-size: 16px;
            margin: 20px 0;
            background: #0a0a0a;
            position: relative;
            z-index: 1;
        }
        .content {
            max-width: 1400px;
            margin: 0 auto;
            position: relative;
            z-index: 1;
        }
        .api-info {
            background: #1a1a1a;
            border: 2px solid #ff0000;
            padding: 20px;
            margin: 20px 0;
            border-radius: 5px;
            font-size: 20px;
            color: #cccccc;
        }
        .endpoint {
            color: #cccccc;
            margin: 5px 0;
            font-size: 20px;
        }
        h2 {
            color: #ff0000;
        }
    </style>
</head>
<body>
    <div class="content">
        <pre class="ascii-art">` + html.EscapeString(getAsciiArtOnly()) + `</pre>
        <div class="title">P A N T E R A</div>
        <div class="subtitle">Dynamic Pricing Engine v1.0.0</div>
        <div class="separator">━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━</div>
        <div class="manifesto" id="manifesto"></div>
        <div class="separator">━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━</div>
        
        <div class="api-info">
            <h2>API Information</h2>
            <p><strong>Name:</strong> Pantera Dynamic Pricing API</p>
            <p><strong>Version:</strong> 1.0.0</p>
            <p><strong>Status:</strong> LIVE</p>
            <p><strong>Description:</strong> Multi-strategy pricing engine</p>
            
            <h2>Endpoints</h2>
            <div class="endpoint">GET  /health - Health check</div>
            <div class="endpoint">GET  /api/v1/rules - Pricing rules</div>
            <div class="endpoint">POST /api/v1/calculate - Calculate pricing</div>
            <div class="endpoint">GET  /api/v1/calculations - Calculation history</div>
            
            <h2>Available Strategies</h2>
            <div class="endpoint">• cost_plus</div>
            <div class="endpoint">• geographic</div>
            <div class="endpoint">• gemstone</div>
        </div>
    </div>
    
    <script>
        // Typing effect for manifesto
        const manifestoHTML = ` + "`" + getManifestoHTML() + "`" + `;
        const manifestoElement = document.getElementById('manifesto');
        let charIndex = 0;
        let tempDiv = document.createElement('div');
        tempDiv.innerHTML = manifestoHTML;
        const textContent = tempDiv.textContent || tempDiv.innerText || '';
        
        function typeText() {
            if (charIndex < textContent.length) {
                let currentHTML = '';
                let plainTextIndex = 0;
                
                // Reconstruct HTML with typing effect
                tempDiv.innerHTML = manifestoHTML;
                const walker = document.createTreeWalker(
                    tempDiv,
                    NodeFilter.SHOW_TEXT,
                    null,
                    false
                );
                
                let node;
                const nodes = [];
                while (node = walker.nextNode()) {
                    nodes.push(node);
                }
                
                let charsAdded = 0;
                for (let node of nodes) {
                    const nodeText = node.textContent;
                    if (charsAdded + nodeText.length <= charIndex) {
                        charsAdded += nodeText.length;
                    } else {
                        const partialText = nodeText.substring(0, charIndex - charsAdded);
                        node.textContent = partialText;
                        charsAdded = charIndex;
                        
                        // Remove all following siblings
                        let next = node.nextSibling;
                        while (next) {
                            const toRemove = next;
                            next = next.nextSibling;
                            toRemove.parentNode.removeChild(toRemove);
                        }
                        
                        // Remove parent's following siblings
                        let parent = node.parentNode;
                        while (parent && parent !== tempDiv) {
                            let next = parent.nextSibling;
                            while (next) {
                                const toRemove = next;
                                next = next.nextSibling;
                                if (toRemove.parentNode) {
                                    toRemove.parentNode.removeChild(toRemove);
                                }
                            }
                            parent = parent.parentNode;
                        }
                        break;
                    }
                }
                
                manifestoElement.innerHTML = tempDiv.innerHTML + '<span class="typing-cursor"></span>';
                charIndex += Math.floor(Math.random() * 3) + 1;
                
                // Random glitch effect
                if (Math.random() > 0.95) {
                    manifestoElement.style.animation = 'glitch 0.1s';
                    setTimeout(() => {
                        manifestoElement.style.animation = '';
                    }, 100);
                }
                
                setTimeout(typeText, 20);
            } else {
                // Finished typing
                manifestoElement.innerHTML = manifestoHTML;
            }
        }
        
        // Start typing after a short delay
        setTimeout(typeText, 500);
    </script>
</body>
</html>
`
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(200, htmlContent)
		} else {
			// JSON response for API clients
			c.JSON(200, gin.H{
				"name":        "Pantera Dynamic Pricing API",
				"version":     "1.0.0",
				"status":      "live",
				"description": "Multi-strategy pricing engine",
				"manifesto": `In the wild, survival belongs to those who can read the terrain before it shifts.
Markets are no different.

Every fluctuation. Every surge. Every unseen variable — hidden in the noise.
Most react. Few anticipate.

Pantera doesn't predict. It hunts.

Forged in code and driven by instinct, Pantera transforms raw data into precision decisions — in milliseconds.

A dynamic pricing engine built not just to calculate, but to sense.

Behind every transaction, a mind that adapts.
Behind every number, an algorithm that learns.

Fast. Silent. Relentless.

This isn't automation. It's evolution.

Pantera — where instinct becomes intelligence.`,
				"endpoints": gin.H{
					"health":       "/health",
					"rules":        "/api/v1/rules",
					"calculate":    "POST /api/v1/calculate",
					"calculations": "/api/v1/calculations",
				},
				"strategies": []string{
					"cost_plus",
					"geographic",
					"gemstone",
				},
			})
		}
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		if err := database.DB.Ping(); err != nil {
			c.JSON(500, gin.H{
				"status":   "unhealthy",
				"database": "disconnected",
				"error":    err.Error(),
			})
			return
		}
		c.JSON(200, gin.H{
			"service":  "base-api",
			"status":   "healthy",
			"version":  "1.0.0",
			"database": "connected",
		})
	})

	// API routes
	routes.SetupPricingRoutes(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("BASE API is now hunting on port %s", port)
	log.Printf("Ready to process dynamic pricing calculations...")
	router.Run(":" + port)
}
