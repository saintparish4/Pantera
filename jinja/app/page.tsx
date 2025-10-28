"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

export default function Home() {
  return (
    <div className="min-h-screen bg-linear-to-br from-white via-zinc-50 to-zinc-100 dark:from-black dark:via-zinc-950 dark:to-zinc-900 relative overflow-hidden">
      {/* Ambient background elements */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-1/2 -right-1/4 w-[800px] h-[800px] bg-linear-to-br from-blue-100/40 via-purple-100/30 to-transparent dark:from-blue-950/20 dark:via-purple-950/10 rounded-full blur-3xl animate-pulse-slow" />
        <div className="absolute -bottom-1/2 -left-1/4 w-[800px] h-[800px] bg-linear-to-tr from-emerald-100/40 via-cyan-100/30 to-transparent dark:from-emerald-950/20 dark:via-cyan-950/10 rounded-full blur-3xl animate-pulse-slower" />
      </div>
      
      <div className="relative z-10">
        <HeroSection />
        <ApiDemoSection />
        <PricingStrategiesSection />
      </div>
    </div>
  );
}

// Hero Section Component
function HeroSection() {
  return (
    <section className="px-6 py-32 text-center">
      <div className="mx-auto max-w-5xl">
        <div className="inline-flex items-center px-4 py-1.5 mb-8 rounded-full bg-white/60 dark:bg-zinc-900/60 backdrop-blur-xl border border-zinc-200/50 dark:border-zinc-800/50 shadow-sm">
          <span className="text-xs font-medium text-zinc-600 dark:text-zinc-400 tracking-wide">
            INTELLIGENT PRICING
          </span>
        </div>
        
        <h1 className="text-6xl md:text-7xl lg:text-8xl font-thin tracking-tight text-zinc-900 dark:text-white mb-8 leading-[0.95]">
          Dynamic<br />
          <span className="font-extralight text-zinc-600 dark:text-zinc-400">Pricing Engine</span>
        </h1>
        
        <p className="mt-8 text-xl md:text-2xl leading-relaxed text-zinc-600 dark:text-zinc-400 font-light max-w-3xl mx-auto">
          Real-time calculations. Intelligent strategies.
          <br />
          <span className="text-zinc-500 dark:text-zinc-500">From cost-plus to gemstone valuation.</span>
        </p>
        
        <div className="mt-16 flex items-center justify-center gap-4">
          <Button 
            size="lg" 
            onClick={() => document.getElementById('demo')?.scrollIntoView({ behavior: 'smooth' })}
            className="px-8 py-6 text-base font-medium rounded-full bg-zinc-900 hover:bg-zinc-800 dark:bg-white dark:hover:bg-zinc-100 dark:text-zinc-900 transition-all duration-300 hover:scale-105 shadow-lg hover:shadow-xl"
          >
            Try Demo
          </Button>
          <Button 
            variant="outline" 
            size="lg" 
            onClick={() => document.getElementById('strategies')?.scrollIntoView({ behavior: 'smooth' })}
            className="px-8 py-6 text-base font-medium rounded-full border-2 border-zinc-300 dark:border-zinc-700 bg-white/60 dark:bg-zinc-900/60 backdrop-blur-xl hover:bg-white dark:hover:bg-zinc-900 transition-all duration-300 hover:scale-105"
          >
            View Strategies
          </Button>
        </div>
      </div>
    </section>
  );
}

// API Demo Section Component
function ApiDemoSection() {
  const [loading, setLoading] = useState(false);
  const [response, setResponse] = useState<Record<string, unknown> | null>(null);
  const [error, setError] = useState<string | null>(null);

  const [formData, setFormData] = useState({
    rule_id: "1",
    quantity: "1",
    demand_level: "1.0",
    location: "",
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setResponse(null);

    try {
      const res = await fetch("http://localhost:8080/api/v1/calculate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          rule_id: parseInt(formData.rule_id),
          quantity: parseInt(formData.quantity),
          demand_level: parseFloat(formData.demand_level),
          location: formData.location || undefined,
        }),
      });

      const data = await res.json();
      
      if (!res.ok) {
        setError(data.error || "Failed to calculate price");
      } else {
        setResponse(data);
      }
    } catch {
      setError("Failed to connect to API. Make sure the backend is running.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <section id="demo" className="px-6 py-24">
      <div className="mx-auto max-w-4xl">
        <div className="text-center mb-16">
          <h2 className="text-5xl md:text-6xl font-thin tracking-tight text-zinc-900 dark:text-white mb-4">
            API Demo
          </h2>
          <p className="text-lg text-zinc-500 dark:text-zinc-500 font-light">
            Experience the power of real-time pricing
          </p>
        </div>
        
        <Card className="border-0 bg-white/80 dark:bg-zinc-900/80 backdrop-blur-xl shadow-2xl">
          <CardHeader className="pb-8">
            <CardTitle className="text-2xl font-light text-zinc-900 dark:text-white">Calculate Price</CardTitle>
            <CardDescription className="text-base text-zinc-500 dark:text-zinc-500">
              Test the pricing API with custom parameters
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-6">
              <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
                <div className="group">
                  <label className="block text-sm font-light text-zinc-600 dark:text-zinc-400 mb-2 tracking-wide">
                    Rule ID
                  </label>
                  <Input
                    type="number"
                    value={formData.rule_id}
                    onChange={(e) => setFormData({ ...formData, rule_id: e.target.value })}
                    required
                    className="h-12 bg-white/50 dark:bg-zinc-950/50 border-zinc-200 dark:border-zinc-800 focus:border-zinc-400 dark:focus:border-zinc-600 transition-all duration-200"
                  />
                </div>
                <div className="group">
                  <label className="block text-sm font-light text-zinc-600 dark:text-zinc-400 mb-2 tracking-wide">
                    Quantity
                  </label>
                  <Input
                    type="number"
                    value={formData.quantity}
                    onChange={(e) => setFormData({ ...formData, quantity: e.target.value })}
                    className="h-12 bg-white/50 dark:bg-zinc-950/50 border-zinc-200 dark:border-zinc-800 focus:border-zinc-400 dark:focus:border-zinc-600 transition-all duration-200"
                  />
                </div>
                <div className="group">
                  <label className="block text-sm font-light text-zinc-600 dark:text-zinc-400 mb-2 tracking-wide">
                    Demand Level
                  </label>
                  <Input
                    type="number"
                    step="0.1"
                    min="0"
                    max="2"
                    value={formData.demand_level}
                    onChange={(e) => setFormData({ ...formData, demand_level: e.target.value })}
                    className="h-12 bg-white/50 dark:bg-zinc-950/50 border-zinc-200 dark:border-zinc-800 focus:border-zinc-400 dark:focus:border-zinc-600 transition-all duration-200"
                  />
                </div>
                <div className="group">
                  <label className="block text-sm font-light text-zinc-600 dark:text-zinc-400 mb-2 tracking-wide">
                    Location
                  </label>
                  <Input
                    type="text"
                    value={formData.location}
                    onChange={(e) => setFormData({ ...formData, location: e.target.value })}
                    placeholder="US, UK, EU..."
                    className="h-12 bg-white/50 dark:bg-zinc-950/50 border-zinc-200 dark:border-zinc-800 focus:border-zinc-400 dark:focus:border-zinc-600 transition-all duration-200"
                  />
                </div>
              </div>
              <Button 
                type="submit" 
                disabled={loading} 
                className="w-full h-12 rounded-full bg-zinc-900 hover:bg-zinc-800 dark:bg-white dark:hover:bg-zinc-100 dark:text-zinc-900 transition-all duration-300 font-light text-base tracking-wide disabled:opacity-50"
              >
                {loading ? "Calculating..." : "Calculate Price"}
              </Button>
            </form>

            {error && (
              <div className="mt-8 rounded-2xl bg-red-50/80 dark:bg-red-950/30 backdrop-blur-sm p-6 border border-red-100 dark:border-red-900/50">
                <p className="text-sm font-light text-red-700 dark:text-red-300">{error}</p>
              </div>
            )}

            {response && (
              <div className="mt-8 rounded-2xl bg-zinc-50/80 dark:bg-zinc-950/50 backdrop-blur-sm p-6 border border-zinc-200 dark:border-zinc-800">
                <h3 className="text-lg font-light mb-4 text-zinc-900 dark:text-zinc-50 tracking-wide">
                  Response
                </h3>
                <pre className="text-sm text-zinc-600 dark:text-zinc-400 overflow-x-auto font-mono leading-relaxed">
                  {JSON.stringify(response, null, 2)}
                </pre>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </section>
  );
}

// Pricing Strategies Section Component
function PricingStrategiesSection() {
  return (
    <section id="strategies" className="px-6 py-32">
      <div className="mx-auto max-w-7xl">
        <div className="text-center mb-20">
          <h2 className="text-5xl md:text-6xl font-thin tracking-tight text-zinc-900 dark:text-white mb-4">
            Pricing Strategies
          </h2>
          <p className="text-lg text-zinc-500 dark:text-zinc-500 font-light">
            Three powerful approaches to intelligent pricing
          </p>
        </div>
        <div className="grid gap-8 lg:grid-cols-3">
          <CostPlusCalculator />
          <GeographicPricingCalculator />
          <GemstoneValuationCalculator />
        </div>
      </div>
    </section>
  );
}

// Cost-Plus Calculator Component
function CostPlusCalculator() {
  const [baseCost, setBaseCost] = useState("100");
  const [markup, setMarkup] = useState("20");
  const [finalPrice, setFinalPrice] = useState<number | null>(null);

  const calculate = () => {
    const cost = parseFloat(baseCost);
    const markupPercent = parseFloat(markup);
    if (!isNaN(cost) && !isNaN(markupPercent)) {
      const price = cost * (1 + markupPercent / 100);
      setFinalPrice(price);
    }
  };

  return (
    <Card className="border-0 bg-white/60 dark:bg-zinc-900/60 backdrop-blur-xl shadow-lg hover:shadow-2xl transition-all duration-500 group">
      <CardHeader className="pb-6">
        <CardTitle className="text-2xl font-light text-zinc-900 dark:text-white">Cost-Plus</CardTitle>
        <CardDescription className="text-sm text-zinc-500 dark:text-zinc-500 font-light">
          Calculate price based on cost and markup percentage
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        <div>
          <label className="block text-xs font-light text-zinc-600 dark:text-zinc-400 mb-2 tracking-wider uppercase">
            Base Cost
          </label>
          <Input
            type="number"
            value={baseCost}
            onChange={(e) => setBaseCost(e.target.value)}
            className="h-11 bg-white/50 dark:bg-zinc-950/50 border-zinc-200 dark:border-zinc-800 transition-all duration-200"
          />
        </div>
        <div>
          <label className="block text-xs font-light text-zinc-600 dark:text-zinc-400 mb-2 tracking-wider uppercase">
            Markup %
          </label>
          <Input
            type="number"
            value={markup}
            onChange={(e) => setMarkup(e.target.value)}
            className="h-11 bg-white/50 dark:bg-zinc-950/50 border-zinc-200 dark:border-zinc-800 transition-all duration-200"
          />
        </div>
        <Button 
          onClick={calculate} 
          className="w-full h-11 rounded-full bg-zinc-900 hover:bg-zinc-800 dark:bg-white dark:hover:bg-zinc-100 dark:text-zinc-900 transition-all duration-300 font-light"
        >
          Calculate
        </Button>
        {finalPrice !== null && (
          <div className="rounded-2xl bg-linear-to-br from-emerald-50 to-teal-50 dark:from-emerald-950/30 dark:to-teal-950/30 p-5 border border-emerald-100 dark:border-emerald-900/50 backdrop-blur-sm">
            <p className="text-xs text-emerald-600 dark:text-emerald-400 font-light tracking-wider uppercase mb-1">
              Final Price
            </p>
            <p className="text-3xl font-thin text-emerald-900 dark:text-emerald-100">
              ${finalPrice.toFixed(2)}
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// Geographic Pricing Calculator Component
function GeographicPricingCalculator() {
  const [basePrice, setBasePrice] = useState("100");
  const [region, setRegion] = useState("US");
  const [finalPrice, setFinalPrice] = useState<number | null>(null);

  const regionMultipliers: Record<string, number> = {
    US: 1.0,
    UK: 1.15,
    EU: 1.20,
    JP: 1.30,
    AU: 1.10,
  };

  const calculate = () => {
    const price = parseFloat(basePrice);
    const multiplier = regionMultipliers[region] || 1.0;
    if (!isNaN(price)) {
      setFinalPrice(price * multiplier);
    }
  };

  return (
    <Card className="border-0 bg-white/60 dark:bg-zinc-900/60 backdrop-blur-xl shadow-lg hover:shadow-2xl transition-all duration-500 group">
      <CardHeader className="pb-6">
        <CardTitle className="text-2xl font-light text-zinc-900 dark:text-white">Geographic Pricing</CardTitle>
        <CardDescription className="text-sm text-zinc-500 dark:text-zinc-500 font-light">
          Adjust prices based on regional multipliers
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        <div>
          <label className="block text-xs font-light text-zinc-600 dark:text-zinc-400 mb-2 tracking-wider uppercase">
            Base Price
          </label>
          <Input
            type="number"
            value={basePrice}
            onChange={(e) => setBasePrice(e.target.value)}
            className="h-11 bg-white/50 dark:bg-zinc-950/50 border-zinc-200 dark:border-zinc-800 transition-all duration-200"
          />
        </div>
        <div>
          <label className="block text-xs font-light text-zinc-600 dark:text-zinc-400 mb-2 tracking-wider uppercase">
            Region
          </label>
          <select
            value={region}
            onChange={(e) => setRegion(e.target.value)}
            className="flex h-11 w-full rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white/50 dark:bg-zinc-950/50 px-4 py-2 text-sm font-light text-zinc-900 dark:text-zinc-100 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-zinc-400 dark:focus:ring-zinc-600"
          >
            <option value="US">United States · 1.0×</option>
            <option value="UK">United Kingdom · 1.15×</option>
            <option value="EU">Europe · 1.20×</option>
            <option value="JP">Japan · 1.30×</option>
            <option value="AU">Australia · 1.10×</option>
          </select>
        </div>
        <Button 
          onClick={calculate} 
          className="w-full h-11 rounded-full bg-zinc-900 hover:bg-zinc-800 dark:bg-white dark:hover:bg-zinc-100 dark:text-zinc-900 transition-all duration-300 font-light"
        >
          Calculate
        </Button>
        {finalPrice !== null && (
          <div className="rounded-2xl bg-linear-to-br from-blue-50 to-indigo-50 dark:from-blue-950/30 dark:to-indigo-950/30 p-5 border border-blue-100 dark:border-blue-900/50 backdrop-blur-sm">
            <p className="text-xs text-blue-600 dark:text-blue-400 font-light tracking-wider uppercase mb-1">
              Regional Price
            </p>
            <p className="text-3xl font-thin text-blue-900 dark:text-blue-100">
              ${finalPrice.toFixed(2)}
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

// Gemstone Valuation Calculator Component
function GemstoneValuationCalculator() {
  const [carat, setCarat] = useState("1.0");
  const [cut, setCut] = useState("Excellent");
  const [clarity, setClarity] = useState("VS1");
  const [finalPrice, setFinalPrice] = useState<number | null>(null);

  const basePricePerCarat = 5000;
  const cutMultipliers: Record<string, number> = {
    Excellent: 1.3,
    "Very Good": 1.15,
    Good: 1.0,
    Fair: 0.85,
  };
  const clarityMultipliers: Record<string, number> = {
    FL: 2.0,
    IF: 1.8,
    VVS1: 1.6,
    VVS2: 1.5,
    VS1: 1.3,
    VS2: 1.2,
    SI1: 1.0,
    SI2: 0.9,
  };

  const calculate = () => {
    const caratWeight = parseFloat(carat);
    const cutMult = cutMultipliers[cut] || 1.0;
    const clarityMult = clarityMultipliers[clarity] || 1.0;
    
    if (!isNaN(caratWeight)) {
      const price = basePricePerCarat * caratWeight * cutMult * clarityMult;
      setFinalPrice(price);
    }
  };

  return (
    <Card className="border-0 bg-white/60 dark:bg-zinc-900/60 backdrop-blur-xl shadow-lg hover:shadow-2xl transition-all duration-500 group">
      <CardHeader className="pb-6">
        <CardTitle className="text-2xl font-light text-zinc-900 dark:text-white">Gemstone Valuation</CardTitle>
        <CardDescription className="text-sm text-zinc-500 dark:text-zinc-500 font-light">
          Calculate diamond price based on 4Cs
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-5">
        <div>
          <label className="block text-xs font-light text-zinc-600 dark:text-zinc-400 mb-2 tracking-wider uppercase">
            Carat Weight
          </label>
          <Input
            type="number"
            step="0.01"
            value={carat}
            onChange={(e) => setCarat(e.target.value)}
            className="h-11 bg-white/50 dark:bg-zinc-950/50 border-zinc-200 dark:border-zinc-800 transition-all duration-200"
          />
        </div>
        <div>
          <label className="block text-xs font-light text-zinc-600 dark:text-zinc-400 mb-2 tracking-wider uppercase">
            Cut Grade
          </label>
          <select
            value={cut}
            onChange={(e) => setCut(e.target.value)}
            className="flex h-11 w-full rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white/50 dark:bg-zinc-950/50 px-4 py-2 text-sm font-light text-zinc-900 dark:text-zinc-100 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-zinc-400 dark:focus:ring-zinc-600"
          >
            <option value="Excellent">Excellent · 1.3×</option>
            <option value="Very Good">Very Good · 1.15×</option>
            <option value="Good">Good · 1.0×</option>
            <option value="Fair">Fair · 0.85×</option>
          </select>
        </div>
        <div>
          <label className="block text-xs font-light text-zinc-600 dark:text-zinc-400 mb-2 tracking-wider uppercase">
            Clarity Grade
          </label>
          <select
            value={clarity}
            onChange={(e) => setClarity(e.target.value)}
            className="flex h-11 w-full rounded-lg border border-zinc-200 dark:border-zinc-800 bg-white/50 dark:bg-zinc-950/50 px-4 py-2 text-sm font-light text-zinc-900 dark:text-zinc-100 transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-zinc-400 dark:focus:ring-zinc-600"
          >
            <option value="FL">FL · 2.0×</option>
            <option value="IF">IF · 1.8×</option>
            <option value="VVS1">VVS1 · 1.6×</option>
            <option value="VVS2">VVS2 · 1.5×</option>
            <option value="VS1">VS1 · 1.3×</option>
            <option value="VS2">VS2 · 1.2×</option>
            <option value="SI1">SI1 · 1.0×</option>
            <option value="SI2">SI2 · 0.9×</option>
          </select>
        </div>
        <Button 
          onClick={calculate} 
          className="w-full h-11 rounded-full bg-zinc-900 hover:bg-zinc-800 dark:bg-white dark:hover:bg-zinc-100 dark:text-zinc-900 transition-all duration-300 font-light"
        >
          Calculate
        </Button>
        {finalPrice !== null && (
          <div className="rounded-2xl bg-linear-to-br from-purple-50 to-pink-50 dark:from-purple-950/30 dark:to-pink-950/30 p-5 border border-purple-100 dark:border-purple-900/50 backdrop-blur-sm">
            <p className="text-xs text-purple-600 dark:text-purple-400 font-light tracking-wider uppercase mb-1">
              Estimated Value
            </p>
            <p className="text-3xl font-thin text-purple-900 dark:text-purple-100">
              ${finalPrice.toFixed(2)}
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
