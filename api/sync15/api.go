package main

/*
	New protocol implementation
	TODO:
		rmapi cache dir
		load cache
			{
				"root":{
					"hash":256,
					"gen":1
				},
				entries:[
					{
						"!name":"from .metatadata",

						"hash":sha256,
						"files": [
							{
								"hash":"",
								"path":"x/y",
								"size":1
							}

						]
					}

				]

		if not exists
			Build a new tree
		else
			merge tree, save

		build ctxtree:
			traverse the tree


		upload new:
			hash each file
			hash [hashes]
			upload blobs

			update index
		download:
			check tree
			find blobs
			get blobs
			zip





*/
