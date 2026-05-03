package data

// Product is a saucedemo inventory item with the kebab-case ID used in selectors.
type Product struct {
	Name string
	// SelectorID is the identifier appended to data-test attributes:
	// e.g., "sauce-labs-backpack" → [data-test="add-to-cart-sauce-labs-backpack"].
	SelectorID string
	Price      string // displayed price string, e.g. "$29.99"
}

// Backpack returns the Sauce Labs Backpack product fixture.
func Backpack() Product {
	return Product{Name: "Sauce Labs Backpack", SelectorID: "sauce-labs-backpack", Price: "$29.99"}
}

// BikeLight returns the Sauce Labs Bike Light product fixture.
func BikeLight() Product {
	return Product{Name: "Sauce Labs Bike Light", SelectorID: "sauce-labs-bike-light", Price: "$9.99"}
}

// BoltTShirt returns the Sauce Labs Bolt T-Shirt product fixture.
func BoltTShirt() Product {
	return Product{Name: "Sauce Labs Bolt T-Shirt", SelectorID: "sauce-labs-bolt-t-shirt", Price: "$15.99"}
}

// FleeceJacket returns the Sauce Labs Fleece Jacket product fixture.
func FleeceJacket() Product {
	return Product{Name: "Sauce Labs Fleece Jacket", SelectorID: "sauce-labs-fleece-jacket", Price: "$49.99"}
}

// Onesie returns the Sauce Labs Onesie product fixture.
func Onesie() Product {
	return Product{Name: "Sauce Labs Onesie", SelectorID: "sauce-labs-onesie", Price: "$7.99"}
}

// AllTheThings returns the Test.allTheThings() T-Shirt (Red) fixture. Special-cased
// SelectorID because the parens in the display name are stripped in the data-test.
func AllTheThings() Product {
	return Product{
		Name:       "Test.allTheThings() T-Shirt (Red)",
		SelectorID: "test.allthethings()-t-shirt-(red)",
		Price:      "$15.99",
	}
}

// AllProducts returns every saucedemo inventory item, in default page order
// (the page's default sort is name A→Z; this list mirrors that order).
func AllProducts() []Product {
	return []Product{
		Backpack(),
		BikeLight(),
		BoltTShirt(),
		FleeceJacket(),
		Onesie(),
		AllTheThings(),
	}
}
