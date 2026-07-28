import "./App.css";
import { ColorModeProvider } from "./lib/color-mode";
import { Header } from "./components/sections/Header";
import { Hero } from "./components/sections/Hero";
import { Features } from "./components/sections/Features";
import { QuickStart } from "./components/sections/QuickStart";
import { Usage } from "./components/sections/Usage";
import { ThemesGallery } from "./components/sections/ThemesGallery";
import { ApiReference } from "./components/sections/ApiReference";
import { Deployment } from "./components/sections/Deployment";
import { Footer } from "./components/sections/Footer";
import { Analytics } from "@vercel/analytics/react";

function App() {
	return (
		<>
			<Analytics />
			<ColorModeProvider>
				<Header />
				<main>
					<Hero />
					<Features />
					<QuickStart />
					<Usage />
					<ThemesGallery />
					<ApiReference />
					<Deployment />
				</main>
				<Footer />
			</ColorModeProvider>
		</>
	);
}

export default App;
