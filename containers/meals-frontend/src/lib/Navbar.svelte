<script lang="ts">
	import { Color } from '$lib/const';
</script>

<div style="--primary-color: {Color.primary}; --secondary-color: {Color.secondary}">
	<nav>
		<a href="/">Home</a>
		<a href="/enable">Enable Meals</a>
	</nav>
</div>

<style>
	nav {
		width: 100%;
		background-color: var(--primary-color);
		overflow: hidden;
	}

	nav a {
		float: left;
		display: block;
		color: white;
		text-align: center;
		text-shadow: black 0.5px 0.5px 0.5px;
		padding: 14px 16px;
		text-decoration: none;
	}

	nav a:hover {
		background-color: var(--secondary-color);
	}
</style>
