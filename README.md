# Advanced EV Trip Planner

Advanced EV Trip Planning & Dynamic Energy Consumption Simulation with GIS Filtering.

## Setup & Running

1. **Install Dependencies** (Go):
   ```bash
   go mod tidy
   ```

2. **Run Server**:
   ```bash
   go run main.go
   ```
   The server will start at `http://localhost:8081`.

3. **Access UI**:
   Open your browser and navigate to `http://localhost:8081`.

## Features

- **Physics-based Simulation**: Calculation includes weight correction and temperature impact.
- **Weather Integration**: Uses Open-Meteo API for real-time temperature data sampled every 50km along the route.
- **GIS Filtering**: Powered by Turf.js to identify Superchargers that meet both distance from origin and proximity to route criteria.
- **Marriott Filter**: Selective display of hotels by brand.
- **Modern Dark UI**: Premium glassmorphism side panel for dynamic parameter adjustment.

## Data Files
- `base.csv`: Vehicle parameters.
- `ec.csv`: Energy consumption matrix (Speed vs Temp).
- `sc.csv`: Tesla Supercharger locations.
- `mariotto.csv`: Marriott Hotel locations.

## License
MIT
