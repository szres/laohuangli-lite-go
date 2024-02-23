<script>
	export let data;
	const caches = data.caches;
	const today = data.today;
	const date = data.date;
	import Today from './today.svelte';
	import Livings from './livings.svelte';
</script>

<div class="flex justify-center w-full h-full">
	<div class="flex flex-col w-full ml-4 mr-4">
		<div class="grid h-1/2 lg:h-2/6 min-h-max content-end">
			<div>
				<Today {today} {date} />
			</div>
		</div>
		<div class="divider mt-4 mb-4"></div>
		<div class="lg:h-3/5 min-h-max">
			<Livings {caches} />
		</div>
	</div>
</div>
