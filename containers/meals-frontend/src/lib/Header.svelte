<script lang="ts">
	import { goto } from '$app/navigation';

	export let title: string;

	function goToHome() {
		goto('/');
	}

	function goToUpdate() {
		goto('/update');
	}
</script>

<h1 style="display: flex; align-items: center; justify-content: flex-start; gap: 1rem;">
	<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />
	<span>{title}</span>
	<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />
</h1>

<div>
	<button on:click={goToHome}> Go Home </button>

	<button on:click={goToUpdate}> Update Meals </button>
</div>
