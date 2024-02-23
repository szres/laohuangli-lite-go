<script>
	export let caches = {};
	import { onMount, onDestroy } from 'svelte';
	import { scale, fade } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	let roller;
	let livingList = [];
	onMount(() => {
		let rollerCount = 0;
		let roll = () => {
			livingList = [livingList[livingList.length - 1], ...livingList];
			livingList = livingList.slice(0, livingList.length - 1);
		};
		for (const k in caches) {
			let one = { id: k, ...caches[k] };
			livingList = [...livingList, one];
		}
		if (livingList.length > 1) {
			for (let index = 0; index < Math.floor(Math.random() * livingList.length); index++) {
				roll();
			}
		}
		roller = setInterval(() => {
			if (rollerCount++ > 20 + Math.floor(Math.random() * 40)) {
				rollerCount = 0;
				if (livingList.length > 1) {
					roll();
				}
			}
		}, 100);
	});
	onDestroy(() => {
		clearInterval(roller);
	});
</script>

<div class="select-none text-3xl text-center font-bold">众生</div>
<div class="flex flex-col mt-2">
	{#if livingList.length > 0}
		{#each livingList.slice(0, 5) as content (content.id)}
			<div
				animate:flip={{ delay: 500 }}
				in:fade={{ delay: 500 }}
				out:scale={{ duration: 1000 }}
				class="mt-2"
			>
				<div class="text-center font-bold text-primary">{content.name}</div>
				<div class="text-center text-xs">{content.result}</div>
			</div>
		{/each}
	{:else}
		<div class="w-full text-sm text-center mt-1">命途未定</div>
	{/if}
</div>
