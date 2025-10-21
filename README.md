# Pantera - Dynamic Pricing Engine API 🚀

**"Stripe for Pricing"** - Developer-focused API for dynamic pricing calculations

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 14+

### Setup (5 minutes)

```bash
# 1. Clone and setup
git clone <your-repo>
cd pantera-api
go mod download

# 2. Setup database
createdb pantera_db

# 3. Configure environment
cp .env.example .env
# Edit .env with your database credentials

# 4. Run the API
go run main.go
```

API will be running at `http://localhost:8080`

---

## API Endpoints

### 💰 Calculate Price (Core Feature)

```bash
POST /api/v1/calculate
```

**Example Request:**
```json
{
  "rule_id": 1,
  "quantity": 5,
  "demand_level": 1.5,
  "competitor_price": 99.99
}
```

**Example Response:**
```json
{
  "price": 105.50,
  "original_price": 110.00,
  "strategy": "demand_based",
  "breakdown": {
    "base_price": 100.00,
    "demand_level": 1.5,
    "demand_adjustment": "50%"
  },
  "calculated_at": "2024-01-15T10:30:00Z"
}
```

---

### 📋 Pricing Rules Management

#### Get All Rules
```bash
GET /api/v1/rules
```

#### Get Single Rule
```bash
GET /api/v1/rules/:id
```

#### Create Rule
```bash
POST /api/v1/rules

{
  "name": "Premium Product Pricing",
  "strategy": "cost_plus",
  "base_price": 100.00,
  "markup_percentage": 25,
  "min_price": 90.00,
  "max_price": 150.00
}
```

#### Update Rule
```bash
PUT /api/v1/rules/:id
```

#### Delete Rule (Soft Delete)
```bash
DELETE /api/v1/rules/:id
```

---

### 📊 Analytics

#### Get Calculation History
```bash
GET /api/v1/calculations?limit=50
```

---

## Pricing Strategies

### 1. Cost-Plus (`cost_plus`)
Simple markup on base price
```json
{
  "strategy": "cost_plus",
  "base_price": 100.00,
  "markup_percentage": 25
}
// Result: $125.00
```

### 2. Demand-Based (`demand_based`)
Adjusts price based on demand levels (0.0 - 2.0)
```json
{
  "strategy": "demand_based",
  "base_price": 100.00,
  "demand_multiplier": 1.0,
  "demand_level": 1.5  // in request
}
// Result: Higher during peak demand
```

### 3. Competitive (`competitive`)
Positions against competitor pricing
```json
{
  "strategy": "competitive",
  "markup_percentage": -5,  // 5% undercut
  "competitor_price": 100.00  // in request
}
// Result: $95.00
```

---

## Testing with cURL

```bash
# 1. Create a pricing rule
curl -X POST http://localhost:8080/api/v1/rules \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Standard Product",
    "strategy": "cost_plus",
    "base_price": 50.00,
    "markup_percentage": 30,
    "min_price": 40.00,
    "max_price": 100.00
  }'

# 2. Calculate a price
curl -X POST http://localhost:8080/api/v1/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "rule_id": 1,
    "quantity": 1
  }'

# 3. Get all rules
curl http://localhost:8080/api/v1/rules

# 4. Get calculation history
curl http://localhost:8080/api/v1/calculations
```

---

## Project Structure

```
pantera-api/
├── main.go                 # Entry point
├── database/
│   └── postgres.go         # DB connection & migrations
├── models/
│   └── pricing.go          # Data models
├── services/
│   └── pricing_engine.go   # Core pricing logic
├── routes/
│   └── pricing.go          # API endpoints
├── .env                    # Environment variables
└── go.mod                  # Dependencies
```

---

---

## Next Steps (Week 1 Remaining)

### Day 3-4: Testing & Seed Data
- [ ] Add example pricing rules via SQL seed file
- [ ] Test all 3 strategies with real calculations
- [ ] Add input validation

### Day 5-6: Documentation
- [ ] Create Postman collection
- [ ] Write integration examples (Node.js, Python, cURL)
- [ ] Deploy to Render

### Day 7: Polish
- [ ] Error handling improvements
- [ ] Rate limiting (optional)
- [ ] API key authentication (basic)

---

## Tech Stack
- **Backend:** Go 1.21 + Gin
- **Database:** PostgreSQL 14
- **Deployment:** Render

---