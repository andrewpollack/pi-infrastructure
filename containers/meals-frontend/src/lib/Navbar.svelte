<script lang="ts">
	import { goto } from '$app/navigation';

	let { title }: { title: string | null } = $props();

	const icon = '<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />';
</script>

<div>
	<ul>
		<li><a href="/" onclick={() => goto('/')}>Home</a></li>
		<li><a href="/update" onclick={() => goto('/')}>Update Meals</a></li>
	</ul>
</div>

{#if title}
	<h1 style="display: flex; align-items: center; justify-content: flex-start; gap: 1rem;">
		{@html icon}
		<span>{title}</span>
		{@html icon}
	</h1>
{/if}

<style>
	ul {
		list-style-type: none;
		margin: 0;
		padding: 0;
		overflow: hidden;
		background-color: #b7b1f2;
	}

	li {
		float: left;
	}

	li a {
		display: block;
		color: white;
		text-align: center;
		padding: 14px 16px;
		text-decoration: none;
	}

	li a:hover {
		background-color: #fdb7ea;
	}
</style>
