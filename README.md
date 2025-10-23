# Pantera - Dynamic Pricing Engine

🚀 **Live API:** https://datos-api-blby.onrender.com

Multi-strategy pricing engine with sophisticated algorithms for e-commerce, SaaS, and luxury goods.

## Quick Demo
```bash
# Health check
curl https://datos-api-blby.onrender.com/health

# Get all pricing strategies
curl https://datos-api-blby.onrender.com/api/v1/rules

# Calculate diamond price
curl -X POST https://datos-api-blby.onrender.com/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "rule_id": 3,
    "gemstone_type": "diamond",
    "carat_weight": 1.5,
    "cut_grade": "excellent",
    "clarity_grade": "VS1",
    "color_grade": "F"
  }'
```

## Pricing Strategies

### 💰 Cost-Plus
Traditional markup pricing for standard products.

### 🌍 Geographic Pricing
Regional market adjustments with purchasing power parity.
- US baseline: $100
- California: $115 (+15%)
- India: $30 (-70% PPP adjustment)

### ⏰ Time-Based Pricing
Dynamic pricing based on time, demand, and seasonality.
- Peak hours (7-9 AM, 5-7 PM): +50%
- Weekend pricing: +20%
- Holiday surges: +80-150%
- Seasonal adjustments (hotels)
- Early bird / last minute pricing (events)
- Optional surge pricing

### 💎 Gemstone Pricing
Luxury goods pricing based on 4Cs plus certification.
- Carat weight with rarity tiers
- Cut quality (Poor to Ideal)
- Clarity grades (I1 to FL)
- Color grades (K-M to D)
- Certification premiums (GIA, IGI, AGS)

## Tech Stack

- **Backend:** Go 1.21 + Gin
- **Database:** PostgreSQL 17 with JSONB
- **Deployment:** Render (auto-scaling)
- **Architecture:** RESTful API

## API Endpoints
```
GET  /health              - Service health
GET  /api/v1/rules        - List strategies
POST /api/v1/calculate    - Calculate price
GET  /api/v1/calculations - History
```

## Example Response
```json
{
  "price": 13702.50,
  "strategy": "gemstone",
  "gemstone_breakdown": {
    "gemstone_type": "diamond",
    "carat_weight": 1.5,
    "total_multiplier": 1.83,
    "price_per_carat": 9134.67
  }
}
```

## Local Development
```bash
git clone https://github.com/YOUR_USERNAME/Pantera.git
cd Pantera/base

export DATABASE_URL="postgresql://localhost:5432/pantera_dev"
export PORT=8080

go run main.go
```

## Deploy to Render
1. Create a PostgreSQL database on Render
2. Create a Web Service and connect it to your database
3. Set environment variables:
   - `DATABASE_URL` (auto-set from database connection)
   - `PORT` (auto-set by Render)
4. Deploy from your GitHub repository

## Features

- Multi-factor pricing algorithms
- JSONB for flexible rule configuration
- Full audit trail with timestamps
- Comprehensive error handling
- Health monitoring
- Production-ready deployment

---

Built with Go + PostgreSQL • Deployed on Render