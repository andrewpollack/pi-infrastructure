<script lang="ts">
	import MealsEmail from '$lib/MealsEmail.svelte';
    import { Icon } from '$lib/const';

	let { data } = $props();
	const { allMeals, allEmails, allExtraItems } =
		data;
</script>

<h1 style="display: flex; align-items: center; justify-content: flex-start; gap: 1rem;">
	{@html Icon}
	<span>Email</span>
	{@html Icon}
</h1>

<div>
	<MealsEmail meals={allMeals} emails={allEmails} extraItems={allExtraItems} />
</div>
