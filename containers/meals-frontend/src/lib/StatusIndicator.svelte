<script lang="ts">
	import { StatusType } from "$lib/types"

	let {
		message = '',
		type = StatusType.SUCCESS,
	}:  { message: string, type: StatusType } = $props();

	let color = $derived.by(() => {
		switch (type) {
			case StatusType.SUCCESS:
				return 'green';
			case StatusType.ERROR:
				return 'red';
			case StatusType.LOADING:
				return 'blue';
		}
	});
</script>

{#if message}
	<div style="color: {color};">
		<p>{message}</p>
	</div>
{/if}
