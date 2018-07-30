#include "network.h"
#include "detection_layer.h"
#include "region_layer.h"
#include "cost_layer.h"
#include "utils.h"
#include "parser.h"
#include "box.h"
#include "image.h"
#include "demo.h"
#include <sys/time.h>

//
#include <linux/videodev2.h>
//#include <libswscale/swscale.h>
#include <errno.h>
#include <sys/ioctl.h>
#include <sys/mman.h>
#include <stdio.h>
#include <fcntl.h>
#include <unistd.h>
//

#define DEMO 1

#ifdef OPENCV

static char **demo_names;
static image **demo_alphabet;
static int demo_classes;

static int imdepth = IPL_DEPTH_8U;
static int imwidth = 480;
static int imheight = 270;
//struct SwsContext *sws_ctx;


static float **probs;
static box *boxes;
static network net;
static image buff [3];
static image buff_letter[3];
static int buff_index = 0;
static CvCapture * cap;
static IplImage  * ipl;
static float fps = 0;
static float demo_thresh = .25;
static float demo_hier = .5;
static int running = 0;

static int demo_delay = 0;
static int demo_frame = 3;
static int demo_detections = 0;
static float **predictions;
static int demo_index = 0;
static int demo_done = 0;
static float *last_avg2;
static float *last_avg;
static float *avg;
double demo_time;

// haxx start
int video_fd = -1;
uint8_t *buffer;
int want_png = 0;

void stdout_image(image im)
{
    int imsize = im.w*im.h*im.c;
    unsigned char *data = calloc(imsize, sizeof(char));
    fprintf(stderr, "writing image to stdout. w:%d, h%d, c:%d, size:%d\n",im.w,im.h,im.c,imsize);
    int i,k;
    for(k = 0; k < im.c; ++k){
        for(i = 0; i < im.w*im.h; ++i){
            data[i*im.c+k] = (unsigned char) (255*im.data[i + k*im.w*im.h]);
        }
    }

    int fp = fopen("/dev/stdout","w");
    fwrite(data, 1, imsize, fp);
    free(data);
}

// this is added to grab raw rgb24 frame from stdin
int grab_frame(image im)
{
	int bsize = (imwidth*imheight*3);
	int cnt = fread(buffer,bsize,1,video_fd);
	int bcnt = cnt*bsize;

	if(bcnt <= 0)
	{
		perror("stdin error");
		return 0;
	}

	// convert to float map
	int i,k;
	for(k = 0; k < im.c; ++k){
		for(i = 0; i < im.w*im.h; ++i){
		im.data[i + k*im.w*im.h] = (float) buffer[i*im.c+k] * (1.f/255.f);
	  }
	}

	return 1;
}

static int xioctl(int fd, int request, void *arg)
{
        int r;
 
        do r = ioctl (fd, request, arg);
        while (-1 == r && EINTR == errno);
 
        return r;
}

int init_mmap()
{
    struct v4l2_requestbuffers req = {0};
    req.count = 1;
    req.type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
    req.memory = V4L2_MEMORY_MMAP;
 
    if (-1 == xioctl(video_fd, VIDIOC_REQBUFS, &req))
    {
        perror("Requesting Buffer");
        return 1;
    }
 
    struct v4l2_buffer buf = {0};
    buf.type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
    buf.memory = V4L2_MEMORY_MMAP;
    buf.index = 0;
    if(-1 == xioctl(video_fd, VIDIOC_QUERYBUF, &buf))
    {
        perror("Querying Buffer");
        return 1;
    }
 
    buffer = mmap (NULL, buf.length, PROT_READ | PROT_WRITE, MAP_SHARED, video_fd, buf.m.offset);
    //printf("Length: %d\nAddress: %p\n", buf.length, buffer);
    //printf("Image Length: %d\n", buf.bytesused);
 
    return 0;
}



double get_wall_time()
{
    struct timeval time;
    if (gettimeofday(&time,NULL)){
        return 0;
    }
    return (double)time.tv_sec + (double)time.tv_usec * .000001;
}

void *detect_in_thread(void *ptr)
{
    running = 1;
    float nms = .4;

    layer l = net.layers[net.n-1];
    float *X = buff_letter[(buff_index+2)%3].data;
    float *prediction = network_predict(net, X);

    memcpy(predictions[demo_index], prediction, l.outputs*sizeof(float));
    mean_arrays(predictions, demo_frame, l.outputs, avg);
    l.output = last_avg2;
    if(demo_delay == 0) l.output = avg;
    if(l.type == DETECTION){
        get_detection_boxes(l, 1, 1, demo_thresh, probs, boxes, 0);
    } else if (l.type == REGION){
        get_region_boxes(l, buff[0].w, buff[0].h, net.w, net.h, demo_thresh, probs, boxes, 0, 0, 0, demo_hier, 1);
    } else {
        error("Last layer must produce detections\n");
    }
    if (nms > 0) do_nms_obj(boxes, probs, l.w*l.h*l.n, l.classes, nms);

    //printf("\033[2J");
    //printf("\033[1;1H");
    //fprintf(stderr,"\nFPS:%.1f\n",fps);
    //fprintf(stderr,"Objects:\n\n");
    image display = buff[(buff_index+2) % 3];
    draw_json_detections(display, demo_detections, demo_thresh, boxes, probs, demo_names, demo_alphabet, demo_classes, buffer, fps, want_png);

    demo_index = (demo_index + 1)%demo_frame;
    running = 0;
    return 0;
}

void *fetch_in_thread(void *ptr)
{
    int status = grab_frame(buff[buff_index]);
    letterbox_image_into(buff[buff_index], net.w, net.h, buff_letter[buff_index]);
    if(status == 0) demo_done = 1;
    return 0;
}

void *display_in_thread(void *ptr)
{
    show_image_cv(buff[(buff_index + 1)%3], "Demo", ipl);
    int c = cvWaitKey(1);
    if (c != -1) c = c%256;
    if (c == 10){
        if(demo_delay == 0) demo_delay = 60;
        else if(demo_delay == 5) demo_delay = 0;
        else if(demo_delay == 60) demo_delay = 5;
        else demo_delay = 0;
    } else if (c == 27) {
        demo_done = 1;
        return 0;
    } else if (c == 82) {
        demo_thresh += .02;
    } else if (c == 84) {
        demo_thresh -= .02;
        if(demo_thresh <= .02) demo_thresh = .02;
    } else if (c == 83) {
        demo_hier += .02;
    } else if (c == 81) {
        demo_hier -= .02;
        if(demo_hier <= .0) demo_hier = .0;
    }
    return 0;
}

void *display_loop(void *ptr)
{
    while(1){
        display_in_thread(0);
    }
}

void *detect_loop(void *ptr)
{
    while(1){
        detect_in_thread(0);
    }
}

//tndemo(cfg, weights, thresh, cam_index, filename, names, classes, frame_skip, prefix, avg, hier_thresh, width, height, fps, fullscreen, png)
void tndemo(char *cfgfile, char *weightfile, float thresh, int cam_index, const char *filename, char **names, int classes, int delay, char *prefix, int avg_frames, float hier, int w, int h, int frames, int fullscreen, int png)
{
    demo_delay = delay;
    demo_frame = avg_frames;
    predictions = calloc(demo_frame, sizeof(float*));
    image **alphabet = load_alphabet();
    demo_names = names;
    demo_alphabet = alphabet;
    demo_classes = classes;
    demo_thresh = thresh;
    demo_hier = hier;
	want_png = png;
	//printf("Demo\n");
    net = parse_network_cfg(cfgfile);
    if(weightfile){
        load_weights(&net, weightfile);
    }
    set_batch_network(&net, 1);
    pthread_t detect_thread;
    pthread_t fetch_thread;

    srand(2222222);

    //sws_ctx = sws_getContext(imwidth, imheight, PIX_FMT_YUV420P, imwidth, imheight, PIX_FMT_BGR24, 2, NULL, NULL, NULL);
    // precreate image that will hold every frame
	if(w) imwidth=w;
	if(h) imheight=h;
    video_fd = fopen("/dev/stdin","r");
    int imsize = (imwidth*imheight*3*4);
    buffer = malloc(imsize);
    memset(buffer,0,sizeof(buffer));

    image im = make_image(imwidth, imheight, 3);

    if(filename){
        fprintf(stderr,"video file: %s\n", filename);
        cap = cvCaptureFromFile(filename);
    }else{
    	if(video_fd==-1) error("Couldn't open stdin.\n");
    }


    layer l = net.layers[net.n-1];
    demo_detections = l.n*l.w*l.h;
    int j;

    avg = (float *) calloc(l.outputs, sizeof(float));
    last_avg  = (float *) calloc(l.outputs, sizeof(float));
    last_avg2 = (float *) calloc(l.outputs, sizeof(float));
    for(j = 0; j < demo_frame; ++j) predictions[j] = (float *) calloc(l.outputs, sizeof(float));

    boxes = (box *)calloc(l.w*l.h*l.n, sizeof(box));
    probs = (float **)calloc(l.w*l.h*l.n, sizeof(float *));
    for(j = 0; j < l.w*l.h*l.n; ++j) probs[j] = (float *)calloc(l.classes+1, sizeof(float));

    grab_frame(im);
    buff[0] = im;
    buff[1] = copy_image(buff[0]);
    buff[2] = copy_image(buff[0]);
    buff_letter[0] = letterbox_image(buff[0], net.w, net.h);
    buff_letter[1] = letterbox_image(buff[0], net.w, net.h);
    buff_letter[2] = letterbox_image(buff[0], net.w, net.h);

    int count = 0;

    demo_time = get_wall_time();

    while(!demo_done){
        buff_index = (buff_index + 1) %3;
        if(pthread_create(&fetch_thread, 0, fetch_in_thread, 0)) error("Thread creation failed");
        if(pthread_create(&detect_thread, 0, detect_in_thread, 0)) error("Thread creation failed");

            if(count % (demo_delay+1) == 0){
                fps = 1./(get_wall_time() - demo_time);
                demo_time = get_wall_time();
                float *swap = last_avg;
                last_avg  = last_avg2;
                last_avg2 = swap;
                memcpy(last_avg, avg, l.outputs*sizeof(float));
            }

        pthread_join(fetch_thread, 0);
        pthread_join(detect_thread, 0);
        ++count;
    }
}

void demo(char *cfgfile, char *weightfile, float thresh, int cam_index, const char *filename, char **names, int classes, int delay, char *prefix, int avg_frames, float hier, int w, int h, int frames, int fullscreen)
{
    demo_delay = delay;
    demo_frame = avg_frames;
    predictions = calloc(demo_frame, sizeof(float*));
    image **alphabet = load_alphabet();
    demo_names = names;
    demo_alphabet = alphabet;
    demo_classes = classes;
    demo_thresh = thresh;
    demo_hier = hier;
    //printf("Demo\n");
    net = parse_network_cfg(cfgfile);
    if(weightfile){
        load_weights(&net, weightfile);
    }
    set_batch_network(&net, 1);
    pthread_t detect_thread;
    pthread_t fetch_thread;

    srand(2222222);

    //sws_ctx = sws_getContext(imwidth, imheight, PIX_FMT_YUV420P, imwidth, imheight, PIX_FMT_BGR24, 2, NULL, NULL, NULL);
    // precreate image that will hold every frame
    video_fd = fopen("/dev/stdin","r");
    int imsize = (imwidth*imheight*3*4);
    buffer = malloc(imsize);
    memset(buffer,0,sizeof(buffer));
    image im = make_image(imwidth, imheight, 3);

    if(filename){
        fprintf(stderr,"video file: %s\n", filename);
        cap = cvCaptureFromFile(filename);
    }else{
	/*
        cap = cvCaptureFromCAM(cam_index);

        if(w){
            cvSetCaptureProperty(cap, CV_CAP_PROP_FRAME_WIDTH, w);
        }
        if(h){
            cvSetCaptureProperty(cap, CV_CAP_PROP_FRAME_HEIGHT, h);
        }
        if(frames){
            cvSetCaptureProperty(cap, CV_CAP_PROP_FPS, frames);
        }
	*/
	/*
	video_fd = open("/dev/video0", O_RDWR);
        if (video_fd == -1)
        {
                perror("Opening video device");
                return;
        }	
	if(init_mmap())
        	return;
	*/
    	if(video_fd==-1) error("Couldn't open stdin.\n");
    }


    layer l = net.layers[net.n-1];
    demo_detections = l.n*l.w*l.h;
    int j;

    avg = (float *) calloc(l.outputs, sizeof(float));
    last_avg  = (float *) calloc(l.outputs, sizeof(float));
    last_avg2 = (float *) calloc(l.outputs, sizeof(float));
    for(j = 0; j < demo_frame; ++j) predictions[j] = (float *) calloc(l.outputs, sizeof(float));

    boxes = (box *)calloc(l.w*l.h*l.n, sizeof(box));
    probs = (float **)calloc(l.w*l.h*l.n, sizeof(float *));
    for(j = 0; j < l.w*l.h*l.n; ++j) probs[j] = (float *)calloc(l.classes+1, sizeof(float));

    //buff[0] = get_image_from_stream(cap);
    grab_frame(im);
    buff[0] = im;
//printf("copy_im_1\n");
    buff[1] = copy_image(buff[0]);
//printf("copy_im_2\n");
    buff[2] = copy_image(buff[0]);
//printf("buff_let1\n");
    buff_letter[0] = letterbox_image(buff[0], net.w, net.h);
//printf("buff_let1\n");
    buff_letter[1] = letterbox_image(buff[0], net.w, net.h);
//printf("buff_let1\n");
    buff_letter[2] = letterbox_image(buff[0], net.w, net.h);
    //printf("creating copy w:%d, h:%d, c:%d\n", buff[0].w,buff[0].h,buff[0].c);
    //ipl = cvCreateImage(cvSize(buff[0].w,buff[0].h), imdepth, buff[0].c);

        //char name[256];
        //sprintf(name, "hello_spec");
        //save_image(buff[0], name);

    int count = 0;
    /*
    if(!prefix){
        cvNamedWindow("Demo", CV_WINDOW_NORMAL); 
        if(fullscreen){
            cvSetWindowProperty("Demo", CV_WND_PROP_FULLSCREEN, CV_WINDOW_FULLSCREEN);
        } else {
            cvMoveWindow("Demo", 0, 0);
            cvResizeWindow("Demo", 1352, 1013);
        }
    }
    */

    demo_time = get_wall_time();

    while(!demo_done){
        buff_index = (buff_index + 1) %3;
        if(pthread_create(&fetch_thread, 0, fetch_in_thread, 0)) error("Thread creation failed");
        if(pthread_create(&detect_thread, 0, detect_in_thread, 0)) error("Thread creation failed");

            if(count % (demo_delay+1) == 0){
                fps = 1./(get_wall_time() - demo_time);
                demo_time = get_wall_time();
                float *swap = last_avg;
                last_avg  = last_avg2;
                last_avg2 = swap;
                memcpy(last_avg, avg, l.outputs*sizeof(float));
            }
	    //stdout_image(buff[(buff_index + 1)%3]);
	/*
	if(count % 2 == 0)
	{
		char name[256];
		sprintf(name, "/var/www/html/predictions");
		save_image(buff[0], name);
	}
	*/

        pthread_join(fetch_thread, 0);
        pthread_join(detect_thread, 0);
        ++count;
    }
}
#else
void demo(char *cfgfile, char *weightfile, float thresh, int cam_index, const char *filename, char **names, int classes, int delay, char *prefix, int avg, float hier, int w, int h, int frames, int fullscreen)
{
    fprintf(stderr, "Demo needs OpenCV for webcam images.\n");
}
#endif

