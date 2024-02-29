<script>
	import { fade } from 'svelte/transition';
	import { onDestroy, onMount } from 'svelte';
	export let data;
	const templatesRegexp = /{{(\w+)}}/g;
	const { templates, entrys, entrysUser } = data;
	let entryCount = 0;
	let entryMetaCount = 0;
	let entryCountUser = 0;
	let entryMetaCountUser = 0;
	onMount(() => {
		for (const key in templates) {
			templates[key].values = [...new Set(templates[key].values)];
			entryMetaCount += templates[key].values.length;
		}
		entryMetaCountUser += entrysUser.length;
		entryMetaCount += entrys.length + entryMetaCountUser;

		for (const item of entrys) {
			entryCount += getDepthOfEntry(item.content);
		}
		for (const item of entrysUser) {
			entryCountUser += getDepthOfEntry(item.content);
		}
		entryCount += entryCountUser;
		console.log('entryMetaCount', entryMetaCount);
		console.log('entryCount', entryCount);
	});

	const getDepthOfEntry = (entry) => {
		let depth = 1;
		[...entry.matchAll(templatesRegexp)].forEach((match) => {
			depth *= templates[match[1]].values.length;
		});
		// console.log(entry, depth)
		return depth;
	};
	const exampleSentences = [
		'吃着{{food}}在{{brand}}店里玩{{game}}',
		'喝着{{drink}}在{{brand}}店里玩{{game}}',
		'在{{wheretoeat}}和{{zodiac}}座的{{student}}一起学习{{programminglanguage}}',
		'{{earlyorlate}}上吃{{food}}，{{earlyorlate}}上喝{{drink}}',
		'不小心把{{drink}}洒在{{gameconsole}}上',
		'穿着全套{{bodywear}}去{{brand}}喝{{coffee}}'
	];
	const exampleIdx = Math.floor(Math.random() * exampleSentences.length);
	let sentence = exampleSentences[exampleIdx];
	let sentenceDepth = 1;
	let templateResult = '-';
	let errTitle = '';
	let errContent = '';
	let error = false;
	const showError = (title, content) => {
		error = true;
		errTitle = title;
		errContent = content;
	};

	let rollerRunning = false;
	let rollerBuilder = () => {
		let id = null;
		const start = () => {
			if (id == null) {
				id = setInterval(() => {
					templateTest(sentence);
				}, 2500);
				rollerRunning = true;
			}
		};
		const stop = () => {
			clearInterval(id);
			id = null;
			rollerRunning = false;
		};
		return { start, stop };
	};
	let roller = rollerBuilder();
	function templateTest(str) {
		if (str.length == 0) {
			return;
		}
		if (str.length > 128) {
			showError('非法词条', '词条长度不能超过128');
			return;
		}
		error = false;
		let templateReplaced = 0;
		const templateReplace = (_, name) => {
			if (!templates.hasOwnProperty(name)) {
				error = true;
				showError('错误', '不存在 {{' + name + '}} 模板');
				return '{{' + name + '}}';
			}
			templateReplaced++;
			return templates[name].values[Math.floor(Math.random() * templates[name].values.length)];
		};
		templateResult = str.replaceAll(templatesRegexp, templateReplace);
		if (templateReplaced == 0) {
			showError('错误', '词条不含任何模板');
		} else if (templateReplaced > 4) {
			showError('错误', '词条使用了超过 4 个模板');
		}
		if (!error) {
			sentenceDepth = getDepthOfEntry(str);
			roller.start();
		} else {
			roller.stop();
		}
	}
	onDestroy(() => {
		roller.stop();
	});
</script>

<div class="flex justify-center">
	<div class="grid content-center max-w-lg lg:max-w-3xl ml-4 mr-4 lg:m-0">
		<div class="flex flex-col justify-center">
			<div class="select-none text-3xl text-center font-bold mt-4">提名助手</div>
			<div class="flex flex-row gap-2 mt-2 lg:pl-4 lg:pr-4">
				<input
					bind:value={sentence}
					type="text"
					placeholder="输入含模板的词条"
					class="grow input input-bordered input-primary"
				/>
				{#if rollerRunning == true}
					<button class="btn btn-primary" on:click={() => roller.stop()}>停止</button>
				{:else}
					<button class="btn btn-primary" on:click={() => templateTest(sentence)}>试试</button>
				{/if}
			</div>
			{#if error}
				<div
					class="flex flex-col justify-center text-center rounded-lg bg-amber-500 text-black py-1 mt-2"
				>
					<h3 class="font-bold text-lg">{errTitle}</h3>
					<p class="py-0">{errContent}</p>
				</div>
			{:else if templateResult.length > 0}
				<!-- <div class="text-center w-full min-h-fit mt-2">
					<span class="mr-2">词条组合数</span><div class="badge badge-secondary">{sentenceDepth}</div>
				</div> -->
				<div class="relative w-full min-h-fit">
					<div
						class="relative select-none w-full text-lg text-center font-bold lg:pl-4 lg:pr-4 text-transparent"
					>
						{templateResult}
					</div>
					{#key templateResult}
						<div
							in:fade
							out:fade
							class="absolute top-0 w-full text-lg text-center font-bold text-primary lg:pl-4 lg:pr-4"
						>
							{templateResult}
						</div>
					{/key}
				</div>
			{/if}
		</div>
		<div class="flex flex-col justify-center text-center">
			<div class="select-none text-3xl font-bold mt-4">词条统计</div>
			<div class="stats stats-vertical lg:stats-horizontal">
				<div class="stat">
					<div class="stat-title">元词条数量</div>
					<div class="stat-value">{entryMetaCount}</div>
				</div>
				<div class="stat">
					<div class="stat-title">词条数量</div>
					<div class="stat-value">{entryCount}</div>
				</div>
				<div class="stat">
					<div class="stat-title">用户提名元词条数量</div>
					<div class="stat-value">{entryMetaCountUser}</div>
				</div>
				<div class="stat">
					<div class="stat-title">用户提名词条数量</div>
					<div class="stat-value">{entryCountUser}</div>
				</div>
			</div>
		</div>
		<div class="mt-2">
			<div class="select-none text-3xl text-center font-bold mt-4">模板列表</div>
			<div class="flex flex-row justify-center flex-wrap gap-2 max-w-max mt-2">
				{#each Object.entries(templates) as [name, content]}
					<div
						class="flex-none has-tooltip min-w-min border rounded-full border-primary overflow-hidden"
					>
						<div
							class="mt-12 tooltip min-w-min max-w-lg rounded-xl p-1 bg-primary flex flex-row justify-center flex-wrap gap-1"
						>
							{#each content.values as v}
								<div class="border border-neutral rounded-md text-neutral text-sm pl-1 pr-1">
									{v}
								</div>
							{/each}
						</div>
						<div class="bg-primary text-neutral w-full text-md font-bold pl-2 pr-2">
							{'{{'}{name}{'}}'}
						</div>
						<div class="text-sm text-center pl-2 pr-2">{content.desc}</div>
					</div>
				{/each}
			</div>
		</div>
		<div class="mt-20"></div>
	</div>
</div>
