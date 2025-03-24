<script lang="ts">
	import { goto } from '$app/navigation';

	let { title }: { title: string } = $props();
</script>

<h1 style="display: flex; align-items: center; justify-content: flex-start; gap: 1rem;">
	<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />
	<span>{title}</span>
	<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />
</h1>

<div>
	<button onclick={() => goto('/')}> Home </button>

	<button onclick={() => goto('/update')}> Update Meals </button>
</div>
