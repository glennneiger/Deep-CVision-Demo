

var is_running = 0;
var canvas_count = 10;
var ctx = new Array(); // will hold our 2d contexts
var playing_banana = 0;
var re_telenor = /telenor/i;
var re_jpeg = /^\/9j\//;
var re_png = /^iVBOR/;
var last_color = '';
var last_prim = '';
var last_size = 0;

function createCanvases() {

	for( var i=0 ; i <= canvas_count ; i++ )
	{
		document.getElementById("canvcont").innerHTML = document.getElementById("canvcont").innerHTML+
		'<div id="canvcont'+i+'" class="canvcont">'+
		'<canvas id="canvas'+i+'" width="100" height="100"></canvas>'+
		'<span class="canvcap" id="cap'+i+'"></span>'+
		'</div>';
	}

	for( var i=0 ; i <= canvas_count ; i++ )
	{
		var this_canvas = document.getElementById("canvas"+i);
		ctx.push(this_canvas.getContext('2d'));
	}

	return ctx;
}

function clearCanvases() {

	for( var i=0 ; i <= canvas_count ; i++ ) {
		var my_canvas = ctx[i].canvas;
		my_canvas.width=100;
		my_canvas.height=100;
		//console.log(i+" "+my_canvas.width);
		ctx[i].clearRect(0, 0, my_canvas.width, my_canvas.canvas);	
		document.getElementById("cap"+i).innerHTML="";
		document.getElementById("canvas"+i).style.border="none";
	}	
}

function drawContext(img,item,ctx,ecaption) {

	var my_canvas = ctx.canvas;
	my_canvas.width=item["box"][2];
	my_canvas.height=item["box"][3];
	ctx.clearRect(0, 0, my_canvas.width, my_canvas.height);
	ecaption.style.marginLeft="-"+(my_canvas.width-8)+"px";
	//ecaption.style.marginTop="-30px";
	ctx.drawImage(img,item["box"][0],item["box"][1],item["box"][2],item["box"][3],0,0,my_canvas.width,my_canvas.height); 
	ecaption.innerHTML=item["class"]+"("+item["pred"]+")";
	if(item["class"] == "banana" && !playing_banana)
	{
		playing_banana = 1;
		document.getElementById('myiframe').src = "https://www.youtube.com/embed/ZYXTZh8CW4E?autoplay=1&showinfo=0&controls=0";
		$('#myiframediv').show(1500);
		setTimeout(function(){
			$('#myiframediv').hide(1000,function() {
				var frame = document.getElementById("myiframe");
				frame.src='';
				playing_banana = 0;
				//frameDoc = frame.contentDocument || frame.contentWindow.document;
				//frameDoc.removeChild(frameDoc.documentElement);
			});
		},30000);
	}

}

function drawPredictions(context, item) {

	context.beginPath();
		context.lineWidth = 1;
		context.strokeStyle=item["color"];
		context.rect(item["box"][0],item["box"][1],item["box"][2],item["box"][3]);
	context.stroke();
}

function info( msg ) {
	document.getElementById("info").innerHTML=msg;
}

function myWebsocketStart() {

	var ws = new WebSocket("ws://"+location.host+":8080/websocket");
	if(!document.getElementById("canvas0"))
		createCanvases();
	var data = {};
	var stream = document.getElementById("stream");
	var stream_ctx = stream.getContext('2d');

	$('#fullscreen').delay( 1000 ).fadeOut( 400 );

	var img = new Image;
	img.onload = function() {

		clearCanvases();
		stream_ctx.drawImage(img,0,0); 
		if(data.hasOwnProperty('meta')) {
			for (var i = 0; i < data.meta.length && i <= canvas_count ; i++) { 
				drawContext(img, data.meta[i], ctx[i], document.getElementById("cap"+i));
				document.getElementById("canvas"+i).style.border="2px solid "+data.meta[i]["color"];
				drawPredictions(stream_ctx,data.meta[i]);
			}
		}
	}

	ws.onmessage = function (evt) {

		if(evt.data[0] != "{")
			return
		data = JSON.parse(evt.data);
		if(data.hasOwnProperty('image')) {
			if( data.image.match(re_jpeg) )
			{
				img.src = "data:image/jpeg;base64,"+data.image;
			}
			else if( data.image.match(re_png) )
			{
				img.src = "data:image/png;base64,"+data.image;
			}
			else
			{
				return;
			}
			is_running = 1;
		}
		else if(data.hasOwnProperty('result')) {
			info(data.result[0].url);
			if(document.getElementById('frame1'))
			{
				document.getElementById('frame1').src = "https://embed.bambuser.com/broadcast/"+data.result[0].vid;
			}
		}
		else if(data.hasOwnProperty('caption')) {
			//info(data.caption);
			$('#fullscreen').text(data.caption);
		    $('#fullscreen').fadeIn( 400 ).delay( 1000 ).fadeOut( 400 );
			if( data.caption.match(re_telenor) )
			{
				 $('#tnlogo').animate({width: "11%", height: "8%", opacity: "0.8"}, 4000);
			}
			var color = data.caption.match(/\br.da?\b/) ? 'red' : 
				data.caption.match(/\bbl..?a?\b/) ? 'blue' :
				data.caption.match(/\bgula?\b/) ? 'yellow' :
				data.caption.match(/\bgr.na?\b/) ? 'green' :
				last_color != '' ? last_color : 'blue';

			var prim = data.caption.match(/cirkel|boll/i) ? 'cirkel' :
				data.caption.match(/fyrkant|rektangel|\bl..?da\b/i) ? 'fyrkant' : 
				last_prim != '' ? last_prim : '';

			var amount = data.caption.match(/\bmycket?\b/) ? 200 :
				data.caption.match(/\bj.tte ?mycket\b/) ? 400 :
				data.caption.match(/\blite\b/) ? 30 :
				data.caption.match(/\bj.tt ?elite\b/) ? 10 : 60;

			var size = data.caption.match(/\bstora?\b/) ? 200 :
				data.caption.match(/\bj.tte ?stora?\b/) ? 400 :
				data.caption.match(/\b(liten|lilla)\b/) ? 60 :
				data.caption.match(/\bj.tte ?(liten|lilla)\b/) ? 30 : 
				last_size != 0 ? last_size : 100;


			var id = color+prim;
			if(data.caption.match(/\bst.rre\b/i))
			{
				size += 50;
			}
			else if(data.caption.match(/\bmindre\b/i))
			{
				size -= 50;
			}

			console.log(data.caption);
			console.log(id);

			var pos = $('#'+id).position();
			last_prim = prim;
			last_color = color;
			last_size = size;
			if( data.caption.match(/\b(hitta|skapa|g.r|rita|lit) ?en|jag vill ha/i) )
			{
				if(prim == 'cirkel')	
				{
					 $('body').append($("<div id='"+id+"' class='primitiv' style='border-radius:400px;background:"+color+";width:"+size+"px;height:"+size+"px'></div>"));
				}
				if(prim == 'fyrkant')	
				{
					 $('body').append($("<div id='"+id+"' class='primitiv' style='background:"+color+";width:"+size+"px;height:"+size+"px'></div>"));
				}
			}
			else if( data.caption.match(/\bta bort|\bradera/i) )
			{
				$('#'+id).remove();

			}
			else if(data.caption.match(/\bh.ger\b/i))
				{
					$('#'+id).animate({left: pos.left+amount},2000);
				}
				else if(data.caption.match(/\bv.nster\b/i))
				{
					$('#'+id).animate({left: pos.left-amount},2000);
				}
				else if(data.caption.match(/\bupp(..)?\b/i))
				{
					$('#'+id).animate({top: pos.top-amount},2000);
				}
				else if(data.caption.match(/\bne[dr](..)?\b/i))
				{
					$('#'+id).animate({top: pos.top+amount},2000);
				}
			else if( size != 0 )
			{
				$('#'+id).animate({width: size+"px", height: size+"px"},2000);	
			}

			//}
			//$('#fullscreen').fadeIn( 100 ).delay( 800 ).fadeOut( 400 );

		}
		else if(data.hasOwnProperty('error')) {
			info(data.error);
		}
	};

	ws.onclose = function(evt) {
		if(evt.code == 1006)
		{
			info("No websocket service detected!");
		}
		is_running = 0;
	};
	is_running = 0;

}

/*
setInterval(function(){ 
	console.log("checking is_running: "+is_running+"\n");
	if( !is_running )
	{
		is_running = 1;
		myWebsocketStart()
		
	}
}, 2000);
*/


