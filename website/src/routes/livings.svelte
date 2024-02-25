<script>
	export let caches = {};
	import { onMount, onDestroy } from 'svelte';
	import { scale, fade } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	let roller;
	let livingPool = [];
	let livingList = [];
	onMount(() => {
		let rollerCount = 0;
		let roll = () => {
			if (livingPool.length > 0) {
				let randIdx = Math.floor(Math.random() * (livingPool.length - 1));
				livingList.unshift(livingPool[randIdx]);
				livingPool.splice(randIdx, 1);
				livingList = livingList;
			}
			if (livingList.length > 5 || (livingPool.length == 0 && livingList.length > 2)) {
				livingPool.push(livingList.pop());
				livingList = livingList;
			}
		};
		for (const k in caches) {
			let one = { id: k, ...caches[k] };
			livingPool.push(one);
		}
		roller = setInterval(() => {
			if (rollerCount++ > 20 + Math.floor(Math.random() * 40)) {
				rollerCount = 0;
				roll();
			}
		}, 100);
	});
	onDestroy(() => {
		clearInterval(roller);
	});
</script>

<div class="select-none text-xl lg:text-3xl text-center font-bold">众生</div>
<div class="flex flex-col mt-2">
	{#if livingList.length > 0}
		{#each livingList as content (content.id)}
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
		<div class="w-full text-sm text-center mt-1">众生命途未定，请稍候片刻</div>
	{/if}
</div>
