<script lang="ts">
	import ExtraItems from '$lib/ExtraItems.svelte';
	import { Icon } from '$lib/const';

	let { data } = $props();
	const { extraItems } = data;
</script>

<h1 style="display: flex; align-items: center; justify-content: flex-start; gap: 1rem;">
	{@html Icon}
	<span> Extra Items </span>
	{@html Icon}
</h1>

<ExtraItems {extraItems} />
