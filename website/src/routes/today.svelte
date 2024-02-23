<script>
	import Daily from './daily.svelte';

	export let today = {
		clothing: {
			positive: '穿衣',
			negative: '裸体'
		},
		food: {
			positive: '吃饭',
			negative: '绝食'
		},
		travel: {
			positive: '走路',
			negative: '爬行'
		}
	};
	let titles = {
		clothing: '穿搭',
		food: '饮食',
		travel: '出行'
	};
	export let date;
</script>

<div class="select-none text-xl lg:text-3xl w-full text-center font-bold">今日</div>
<div class="w-full text-xs lg:text-sm text-center">{date}</div>
<div class="flex justify-center">
	<div class="flex flex-col mt-2 inline-flex">
		{#each Object.entries(today) as [k, v]}
			<Daily title={titles[k]} positive={v.positive} negative={v.negative} />
		{/each}
	</div>
</div>
