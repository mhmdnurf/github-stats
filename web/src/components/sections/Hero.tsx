import { Container } from "../ui/Container";
import { LinkButton } from "../ui/Button";
import { statsCardUrl, languagesCardUrl, SITE } from "../../lib/config";
import { useColorMode } from "../../lib/use-color-mode";

export function Hero() {
	const { mode } = useColorMode();
	const cardTheme = mode === "dark" ? "default" : "light";

	return (
		<section id="top" className="hero-section">
			<Container>
				<h1>Turn GitHub activity into embeddable SVG stat cards</h1>
				<p className="hero-lede">
					GitHub Stats periodically snapshots your profile through a scheduled
					refresh job and serves fast, native SVG cards &mdash; no client-side
					JavaScript, no browser rendering, no calls to GitHub on the request
					path.
				</p>
				<div className="hero-actions">
					<LinkButton href="#quick-start" variant="primary">
						Get started
					</LinkButton>
					<LinkButton
						href={SITE.repoUrl}
						target="_blank"
						rel="noreferrer"
						variant="secondary"
					>
						View source
					</LinkButton>
				</div>
				<div className="hero-preview">
					<img
						src={statsCardUrl(cardTheme)}
						alt="GitHub statistics card preview"
						loading="lazy"
					/>
					<img
						src={languagesCardUrl(cardTheme)}
						alt="Most used languages card preview"
						loading="lazy"
					/>
				</div>
			</Container>
		</section>
	);
}
