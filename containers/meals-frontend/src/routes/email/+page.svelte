<script lang="ts">
	import MealsEmail from '$lib/MealsEmail.svelte';

	let { data } = $props();
	const { allMeals, allEmails, allExtraItems } = data;
</script>

<h1 style="display: flex; align-items: center; justify-content: flex-start; gap: 1rem;">
	<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />
	<span>Email</span>
	<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />
</h1>

<div>
	<MealsEmail meals={allMeals} emails={allEmails} extraItems={allExtraItems} />
</div>
