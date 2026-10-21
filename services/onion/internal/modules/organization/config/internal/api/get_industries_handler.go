package api

type GetIndustriesHandler struct{}

type GetIndustriesRequest struct{}

type IndustryTypePayload struct {
	Slug  string `json:"slug"`
	Label string `json:"label"`
	Icon  string `json:"icon"`
}

type IndustryGroupPayload struct {
	Category string                `json:"category"`
	Types    []IndustryTypePayload `json:"types"`
}

type GetIndustriesResponse struct {
	Groups []IndustryGroupPayload `json:"groups"`
}

var industryGroups = []IndustryGroupPayload{
	{
		Category: "Food & Beverage",
		Types: []IndustryTypePayload{
			{Slug: "restaurant", Label: "Restaurant", Icon: "🍽️"},
			{Slug: "cafe", Label: "Café", Icon: "☕"},
			{Slug: "bar", Label: "Bar", Icon: "🍺"},
			{Slug: "food_truck", Label: "Food Truck", Icon: "🚚"},
			{Slug: "bakery", Label: "Bakery", Icon: "🥐"},
			{Slug: "juice_bar", Label: "Juice Bar", Icon: "🥤"},
			{Slug: "catering", Label: "Catering", Icon: "🍱"},
			{Slug: "fast_food", Label: "Fast Food", Icon: "🍔"},
			{Slug: "pizzeria", Label: "Pizzeria", Icon: "🍕"},
			{Slug: "sushi", Label: "Sushi Restaurant", Icon: "🍣"},
			{Slug: "steakhouse", Label: "Steakhouse", Icon: "🥩"},
			{Slug: "seafood", Label: "Seafood Restaurant", Icon: "🦞"},
			{Slug: "vegetarian", Label: "Vegetarian Restaurant", Icon: "🥗"},
			{Slug: "vegan", Label: "Vegan Restaurant", Icon: "🌱"},
			{Slug: "buffet", Label: "Buffet", Icon: "🍛"},
			{Slug: "ice_cream", Label: "Ice Cream Shop", Icon: "🍦"},
			{Slug: "dessert", Label: "Dessert Shop", Icon: "🍰"},
			{Slug: "bubble_tea", Label: "Bubble Tea", Icon: "🧋"},
			{Slug: "brewery", Label: "Brewery", Icon: "🍻"},
			{Slug: "winery", Label: "Winery", Icon: "🍷"},
			{Slug: "distillery", Label: "Distillery", Icon: "🥃"},
			{Slug: "ghost_kitchen", Label: "Ghost Kitchen", Icon: "👻"},
			{Slug: "meal_prep", Label: "Meal Prep", Icon: "🥡"},
			{Slug: "food_delivery", Label: "Food Delivery", Icon: "🛵"},
		},
	},
	{
		Category: "Retail",
		Types: []IndustryTypePayload{
			{Slug: "retail_clothing", Label: "Clothing Store", Icon: "👗"},
			{Slug: "retail_grocery", Label: "Grocery Store", Icon: "🛒"},
			{Slug: "retail_pharmacy", Label: "Pharmacy", Icon: "💊"},
			{Slug: "retail_electronics", Label: "Electronics Store", Icon: "📱"},
			{Slug: "retail_gifts", Label: "Gift Shop", Icon: "🎁"},
			{Slug: "retail_books", Label: "Bookstore", Icon: "📚"},
			{Slug: "retail_jewelry", Label: "Jewelry Store", Icon: "💍"},
			{Slug: "retail_shoes", Label: "Shoe Store", Icon: "👟"},
			{Slug: "retail_sports", Label: "Sports Store", Icon: "⚽"},
			{Slug: "retail_furniture", Label: "Furniture Store", Icon: "🛋️"},
			{Slug: "retail_hardware", Label: "Hardware Store", Icon: "🔧"},
			{Slug: "retail_toys", Label: "Toy Store", Icon: "🧸"},
			{Slug: "retail_pets", Label: "Pet Store", Icon: "🐾"},
			{Slug: "retail_cosmetics", Label: "Cosmetics Store", Icon: "💄"},
			{Slug: "retail_flowers", Label: "Flower Shop", Icon: "💐"},
			{Slug: "retail_music", Label: "Music Store", Icon: "🎸"},
			{Slug: "retail_art", Label: "Art Gallery & Shop", Icon: "🎨"},
			{Slug: "retail_vintage", Label: "Vintage & Thrift Store", Icon: "🪣"},
			{Slug: "convenience_store", Label: "Convenience Store", Icon: "🏪"},
			{Slug: "supermarket", Label: "Supermarket", Icon: "🏬"},
		},
	},
	{
		Category: "Health & Beauty",
		Types: []IndustryTypePayload{
			{Slug: "hair_salon", Label: "Hair Salon", Icon: "💇"},
			{Slug: "barbershop", Label: "Barbershop", Icon: "💈"},
			{Slug: "spa", Label: "Spa", Icon: "🧖"},
			{Slug: "gym", Label: "Gym & Fitness Center", Icon: "🏋️"},
			{Slug: "nail_salon", Label: "Nail Salon", Icon: "💅"},
			{Slug: "medical_clinic", Label: "Medical Clinic", Icon: "🏥"},
			{Slug: "dental", Label: "Dental Clinic", Icon: "🦷"},
			{Slug: "optician", Label: "Optician", Icon: "👓"},
			{Slug: "physiotherapy", Label: "Physiotherapy", Icon: "🩺"},
			{Slug: "massage", Label: "Massage Therapy", Icon: "💆"},
			{Slug: "yoga", Label: "Yoga Studio", Icon: "🧘"},
			{Slug: "pilates", Label: "Pilates Studio", Icon: "🤸"},
			{Slug: "crossfit", Label: "CrossFit Box", Icon: "🏅"},
			{Slug: "martial_arts", Label: "Martial Arts", Icon: "🥋"},
			{Slug: "tattoo", Label: "Tattoo Studio", Icon: "🖊️"},
			{Slug: "piercing", Label: "Piercing Studio", Icon: "💎"},
			{Slug: "tanning", Label: "Tanning Salon", Icon: "☀️"},
			{Slug: "beauty_school", Label: "Beauty School", Icon: "🎓"},
			{Slug: "dermatology", Label: "Dermatology", Icon: "🧴"},
			{Slug: "cosmetic_surgery", Label: "Cosmetic Surgery", Icon: "✨"},
		},
	},
	{
		Category: "Hospitality",
		Types: []IndustryTypePayload{
			{Slug: "hotel", Label: "Hotel", Icon: "🏨"},
			{Slug: "hostel", Label: "Hostel", Icon: "🛏️"},
			{Slug: "resort", Label: "Resort", Icon: "🌴"},
			{Slug: "guesthouse", Label: "Guesthouse", Icon: "🏡"},
			{Slug: "motel", Label: "Motel", Icon: "🛣️"},
			{Slug: "serviced_apartment", Label: "Serviced Apartment", Icon: "🏢"},
			{Slug: "villa_rental", Label: "Villa Rental", Icon: "🏖️"},
			{Slug: "camping", Label: "Camping & Glamping", Icon: "⛺"},
			{Slug: "airbnb_management", Label: "Short-Term Rental Management", Icon: "🔑"},
		},
	},
	{
		Category: "Professional Services",
		Types: []IndustryTypePayload{
			{Slug: "law_firm", Label: "Law Firm", Icon: "⚖️"},
			{Slug: "accounting", Label: "Accounting & Bookkeeping", Icon: "📊"},
			{Slug: "consulting", Label: "Business Consulting", Icon: "💼"},
			{Slug: "marketing_agency", Label: "Marketing Agency", Icon: "📣"},
			{Slug: "pr_agency", Label: "PR Agency", Icon: "📰"},
			{Slug: "architecture", Label: "Architecture Firm", Icon: "🏛️"},
			{Slug: "engineering", Label: "Engineering Firm", Icon: "⚙️"},
			{Slug: "recruitment", Label: "Recruitment Agency", Icon: "🤝"},
			{Slug: "insurance", Label: "Insurance Agency", Icon: "🛡️"},
			{Slug: "financial_advisor", Label: "Financial Advisor", Icon: "💹"},
		},
	},
	{
		Category: "Technology",
		Types: []IndustryTypePayload{
			{Slug: "software_dev", Label: "Software Development", Icon: "💻"},
			{Slug: "it_support", Label: "IT Support & Services", Icon: "🖥️"},
			{Slug: "cybersecurity", Label: "Cybersecurity", Icon: "🔐"},
			{Slug: "data_analytics", Label: "Data Analytics", Icon: "📈"},
			{Slug: "ai_company", Label: "AI Company", Icon: "🤖"},
			{Slug: "web_design", Label: "Web Design", Icon: "🌐"},
			{Slug: "mobile_dev", Label: "Mobile Development", Icon: "📲"},
			{Slug: "game_dev", Label: "Game Development", Icon: "🎮"},
			{Slug: "saas", Label: "SaaS", Icon: "☁️"},
			{Slug: "cloud_services", Label: "Cloud Services", Icon: "🌩️"},
		},
	},
	{
		Category: "Education",
		Types: []IndustryTypePayload{
			{Slug: "school", Label: "School", Icon: "🏫"},
			{Slug: "university", Label: "University & College", Icon: "🎓"},
			{Slug: "tutoring", Label: "Tutoring Center", Icon: "📝"},
			{Slug: "language_school", Label: "Language School", Icon: "🗣️"},
			{Slug: "driving_school", Label: "Driving School", Icon: "🚗"},
			{Slug: "music_school", Label: "Music School", Icon: "🎵"},
			{Slug: "art_school", Label: "Art School", Icon: "🖼️"},
			{Slug: "cooking_class", Label: "Cooking Class", Icon: "👨‍🍳"},
			{Slug: "online_courses", Label: "Online Courses", Icon: "💡"},
			{Slug: "daycare", Label: "Daycare & Childcare", Icon: "👶"},
		},
	},
	{
		Category: "Entertainment",
		Types: []IndustryTypePayload{
			{Slug: "cinema", Label: "Cinema", Icon: "🎬"},
			{Slug: "theater", Label: "Theater", Icon: "🎭"},
			{Slug: "concert_venue", Label: "Concert Venue", Icon: "🎤"},
			{Slug: "nightclub", Label: "Nightclub", Icon: "🪩"},
			{Slug: "escape_room", Label: "Escape Room", Icon: "🔓"},
			{Slug: "gaming_cafe", Label: "Gaming Café", Icon: "🕹️"},
			{Slug: "bowling", Label: "Bowling Alley", Icon: "🎳"},
			{Slug: "paintball", Label: "Paintball", Icon: "🎯"},
			{Slug: "trampoline_park", Label: "Trampoline Park", Icon: "🤸"},
			{Slug: "amusement_park", Label: "Amusement Park", Icon: "🎡"},
			{Slug: "mini_golf", Label: "Mini Golf", Icon: "⛳"},
			{Slug: "go_kart", Label: "Go-Kart Track", Icon: "🏎️"},
		},
	},
	{
		Category: "Transportation",
		Types: []IndustryTypePayload{
			{Slug: "taxi", Label: "Taxi Service", Icon: "🚕"},
			{Slug: "rideshare", Label: "Rideshare", Icon: "🚖"},
			{Slug: "car_rental", Label: "Car Rental", Icon: "🚘"},
			{Slug: "trucking", Label: "Trucking & Logistics", Icon: "🚛"},
			{Slug: "moving_company", Label: "Moving Company", Icon: "📦"},
			{Slug: "courier", Label: "Courier & Delivery", Icon: "🛵"},
			{Slug: "bike_rental", Label: "Bike & Scooter Rental", Icon: "🚲"},
			{Slug: "travel_agency", Label: "Travel Agency", Icon: "✈️"},
			{Slug: "airline", Label: "Airline", Icon: "🛫"},
			{Slug: "shipping", Label: "Shipping & Freight", Icon: "🚢"},
		},
	},
	{
		Category: "Finance",
		Types: []IndustryTypePayload{
			{Slug: "bank", Label: "Bank", Icon: "🏦"},
			{Slug: "credit_union", Label: "Credit Union", Icon: "💳"},
			{Slug: "forex", Label: "Forex Exchange", Icon: "💱"},
			{Slug: "crypto_exchange", Label: "Crypto Exchange", Icon: "₿"},
			{Slug: "payment_processing", Label: "Payment Processing", Icon: "💰"},
			{Slug: "microfinance", Label: "Microfinance", Icon: "🤲"},
		},
	},
	{
		Category: "Real Estate",
		Types: []IndustryTypePayload{
			{Slug: "real_estate_agency", Label: "Real Estate Agency", Icon: "🏠"},
			{Slug: "property_management", Label: "Property Management", Icon: "🗝️"},
			{Slug: "construction", Label: "Construction", Icon: "🏗️"},
			{Slug: "interior_design", Label: "Interior Design", Icon: "🛋️"},
			{Slug: "cleaning_service", Label: "Cleaning Service", Icon: "🧹"},
			{Slug: "landscaping", Label: "Landscaping", Icon: "🌿"},
		},
	},
	{
		Category: "Manufacturing",
		Types: []IndustryTypePayload{
			{Slug: "food_production", Label: "Food Production", Icon: "🏭"},
			{Slug: "textile", Label: "Textile & Apparel", Icon: "🧵"},
			{Slug: "furniture_maker", Label: "Furniture Maker", Icon: "🪑"},
			{Slug: "printing", Label: "Printing & Publishing", Icon: "🖨️"},
			{Slug: "packaging", Label: "Packaging", Icon: "📦"},
			{Slug: "electronics_manufacturing", Label: "Electronics Manufacturing", Icon: "🔌"},
		},
	},
	{
		Category: "Agriculture",
		Types: []IndustryTypePayload{
			{Slug: "farm", Label: "Farm", Icon: "🌾"},
			{Slug: "greenhouse", Label: "Greenhouse", Icon: "🌿"},
			{Slug: "livestock", Label: "Livestock", Icon: "🐄"},
			{Slug: "organic_produce", Label: "Organic Produce", Icon: "🥬"},
			{Slug: "fishery", Label: "Fishery & Aquaculture", Icon: "🐟"},
			{Slug: "beekeeping", Label: "Beekeeping", Icon: "🐝"},
		},
	},
	{
		Category: "Other",
		Types: []IndustryTypePayload{
			{Slug: "events_venue", Label: "Events Venue", Icon: "🎪"},
			{Slug: "photography", Label: "Photography Studio", Icon: "📷"},
			{Slug: "videography", Label: "Videography", Icon: "🎥"},
			{Slug: "charity", Label: "Charity & Non-Profit", Icon: "❤️"},
			{Slug: "religious_org", Label: "Religious Organization", Icon: "🕌"},
			{Slug: "government", Label: "Government Agency", Icon: "🏛️"},
			{Slug: "coworking", Label: "Coworking Space", Icon: "🖥️"},
			{Slug: "storage", Label: "Storage Facility", Icon: "🗄️"},
			{Slug: "laundry", Label: "Laundry & Dry Cleaning", Icon: "👕"},
			{Slug: "auto_repair", Label: "Auto Repair Shop", Icon: "🔧"},
			{Slug: "other", Label: "Other", Icon: "🏢"},
		},
	},
}
