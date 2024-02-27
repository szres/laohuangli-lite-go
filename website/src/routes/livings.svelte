<script>
	export let caches = {};
	import { onMount, onDestroy } from 'svelte';
	import { scale, fade } from 'svelte/transition';
	import { flip } from 'svelte/animate';
	let roller;
	let livingPool = [];
	let livingList = [];
	let livingListRemove = 0;
	function poolCreateFrom(cache) {
		livingPool.splice(0, livingPool.length);
		for (const k in cache) {
			let one = { id: k, ...cache[k] };
			if (!livingPool.hasOwnProperty(k)) {
				livingPool.push(one);
			}
		}
		livingListRemove = livingList.length;
	}
	const shuffle = (array) => {
		for (let i = array.length - 1; i > 0; i--) {
			const j = Math.floor(Math.random() * (i + 1));
			[array[i], array[j]] = [array[j], array[i]];
		}
		return array;
	};
	$: poolCreateFrom(caches);
	onMount(() => {
		let rollerCount = 0;
		let roll = () => {
			if (livingPool.length > 0) {
				livingPool = shuffle(livingPool);
				let exist = false;
				for (const item of livingList) {
					if (item.id === livingPool[0]?.id) {
						exist = true;
						break;
					}
				}
				if (!exist) {
					livingList = [livingPool.shift(), ...livingList];
				}
			}
			if (
				livingList.length > 5 ||
				(livingPool.length == 0 && livingList.length > 2) ||
				livingListRemove > 0
			) {
				let item = livingList.pop();
				if (livingListRemove > 0) {
					livingListRemove--;
				} else {
					livingPool.push(item);
				}
				livingList = livingList;
			}
		};
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
	export async function load({ parent }) {
		const { a, b } = await parent();
		return { c: a + b };
	}
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
