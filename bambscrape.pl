#!/usr/bin/perl


use strict;
use Data::Dumper;
use JSON::XS qw(decode_json);

our $DEBUG = 0;

# Grab video index and continue if live
my $res = qx{ curl 'https://bambuser.com/xhr-api/index.php?username=r00tz&sort=created&access_mode=0%2C1%2C2&limit=12&_strict=1&method=broadcast&format=json&_=1502393968952'  -H 'Accept-Encoding: gzip, deflate, br' -H 'Content-Type: application/x-www-form-urlencoded; charset=UTF-8' -H 'Accept: application/json, text/javascript, */*; q=0.01' -H 'X-Requested-With: XMLHttpRequest' -H 'Connection: keep-alive' -H 'Referer: https://bambuser.com/channel/r00tz' --compressed 2>/dev/null };
print "$res\n\n" if $DEBUG;

my $json = decode_json($res);
my $vid = $json->{result}[0]{vid} if ref $json->{result}[0] eq 'HASH';
print "$vid\n\n" if $DEBUG;
exit if !$vid || $json->{result}[0]{type} ne 'live';


# Grab resourceUri including signed session id 
exit unless qx{ curl "https://bambuser.com/v/$vid" 2>/dev/null } =~ /resourceUri = '(.+?)'/;
my $resource = $1;
print "$resource\n\n" if $DEBUG;


# Grab and traspose cdn streaming mp4 url
$res = qx{curl "https://cdn.bambuser.net/contentRequests" -H "Content-Type: application/json" -H "Accept: application/vnd.bambuser.cdn.v1+json" -H "X-Iris-ApplicationId: BAMLPazITw28Uj9vnHeRQX" -H "X-Bambuser-ClientVersion: com.bambuser.BambuserVideoJS/0.6.1 bambuser-cdn-client-js/0.0.1" -H "X-Bambuser-ClientPlatform: html5" --data-binary '{"resourceUri":"$resource","broadcastState":"viewable","criteria":[{"preset":"mp4-h264","protocol":"wss"},{"preset":"hls"}]}' --compressed 2>/dev/null};
exit unless $res =~ /^{/;
my $cdn = decode_json($res);
print Dumper $cdn if $DEBUG;
my $stream = $cdn->{match}{url} =~ s#wss://(?:relay.*?bambuser[.]net/)?#http://#r;

print "$stream\n" if $cdn->{match}{live};

