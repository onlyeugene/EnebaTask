# Design reference

Source: eneba.com homepage, Chrome DevTools CSS Overview + Inspect (Nov 2026).
Screenshots in this folder are local reference only (gitignored).
Items marked (check) are inferred from screenshots, not confirmed in DevTools.

## Colors

### Purples (main brand)
- Page background: #4618ac
- Lighter purple (homepage section bands, alternating with #4618ac): #5825cc
- Dark purple (nav bar, darker sections): #351183 (check)
- Darkest purple (card info area, deep surfaces): #1f0a4d (check)
- Border purple: #3c1395
- Light purple (hover, accents): #7a4edf (check)

### Accents
- Yellow (main CTA buttons, highlights, countdown): #fad318
- Mint (cashback badge background, product card border): #63e3c2
- Lime green (discount %, cashback amount): #84e916
- Teal green (region label, e.g. EUROPE, GLOBAL): #23c299
- Pink (region label, e.g. UNITED STATES, UNITED KINGDOM): #e76a8c
- Red (errors, alerts, cart count badge): #eb002e
- Red-pink (sale/alert surfaces): #d32252 (check)
- Dark red (error borders): #99001f

### Text
- Primary: #ffffff
- Primary soft (85% white): #ffffffd9
- Secondary: #dcd8e6
- Muted (old prices, labels): #b3aac9
- More muted (meta text): #9385b2
- Dark text on light/yellow buttons: #1f0a4d (check)

### Neutrals and overlays
- White surface: #ffffff
- Light gray surface: #efefef
- Subtle overlay / divider (10% white): #ffffff1a

## Typography
- Font family (original): Metropolis, fallbacks Arial, Helvetica, sans-serif
- Font family (using): Metropolis via @fontsource/metropolis (confirm package exists), same fallbacks
- Ignore Open Sans (only 1 occurrence, likely a third-party widget)

### Sizes (most used first)
- 12px: small labels, meta text (most common)
- 16px: body
- 10px: tiny badges
- 15px: secondary body
- 14px: buttons, inputs
- 22px: section headings
- 30px: large prices, page headings

### Weights
- 800: headings, prices, titles (most common, very bold look)
- 400: body
- 600: semi-bold labels
- 700: buttons
- 900: rare, skip

### Line heights
- 12px text: 12px to 15px
- 16px text: 24px
- 22px text: 22px
- 30px text: 37.5px (1.25)

## Breakpoints
- 400px: small phones
- 600px: phones / tablets
- 930px: tablets / desktop
- 1200px: wide desktop
- Separate hover styles for (hover: hover) vs touch devices

## Shape and spacing
- Max content width: 1200px (centered)
- Section spacing: 40px below each section
- Border radius: 0 everywhere observed (cards, search bar, See all button). Square corners are part of the look.

### Product grid
- 5 columns of 224px, 20px column gap, 20px row gap
- Fewer columns at smaller breakpoints

### Product card
- Outer: 224px wide, about 458px tall, 1px solid #63e3c2 border, no padding
- Image area: 222 x 250px, transparent background
- Info area: dark purple background (#1f0a4d, check), about 20px padding (estimate)

### Header search bar
- 550 x 57px
- 1px solid #ffffff border, no fill (transparent on header purple)
- Padding: 0 10px
- Line height: 24px
- Search icon on the left, placeholder text in muted white

### Buttons
- See all: outlined, 1px solid white border, transparent background, white bold text, square corners
- Primary CTA: solid #fad318, dark text, bold, square corners

## Components noted

### Header
- Background #4618ac
- Logo left, large search input center, then language/currency, wishlist heart,
  cart icon with red count badge, Log in | Register on the right
- Below: darker nav bar with hamburger "Categories" and text links

### Hero
- Wide main banner on the left, 3 narrower promo tiles on the right, edge to edge

### Category row
- Horizontal scroll of icon + label items (white outline icons, bold labels)
- Arrow buttons left and right

### Section heading
- Large bold white title (e.g. "Recommended for you") with a small info icon
- "See all" outlined button centered below the grid

### Product card
- Cover image on top
- Platform strip at the bottom of the image (icon + "Steam" / "PSN")
- Mint "CASHBACK" pill overlapping the image bottom left
- Optional "GIFT CARD" value tag in the top right corner of the image
- Dark purple info area below:
  - Title: white, bold, 2 lines max with ellipsis
  - Region: uppercase, bold, colored (teal for EUROPE/GLOBAL, pink for UNITED STATES/UNITED KINGDOM)
  - "From" + old price with strikethrough (muted) + discount % (lime)
  - Current price: large (about 30px), white, weight 800, small info icon after it
  - "Cashback: €X.XX" in lime
- 1px mint border around the card, square corners

### Badges
- Cashback: mint background, dark text, uppercase, icon on the left
- Gift card value: small uppercase tag in the top corner of the image
- Cart count: small red circle with white number

## Branding
- Use my own app name and logo, not Eneba's
- Footer note: "Demo project, not affiliated with Eneba"
- No copyrighted images; product covers are generated SVGs from the seed script