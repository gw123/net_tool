//发送端
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <errno.h>
#include <sys/socket.h>
#include <sys/ioctl.h>
#include <sys/types.h>
#include <netinet/in.h>
#include <netinet/ip.h>
#include <netinet/if_ether.h>
#include <net/if_arp.h>
#include <netpacket/packet.h>
#include <net/if.h>
#include <net/ethernet.h>

#define BUFLEN 42

int main(int argc,char** argv){
    int skfd,n;
    char buf[BUFLEN]={0};
    struct ether_header *eth;
    struct ether_arp *arp;
    struct sockaddr_ll toaddr;
    struct in_addr targetIP,srcIP;
    struct ifreq ifr;

    unsigned char src_mac[ETH_ALEN]={0};
    unsigned char dst_mac[ETH_ALEN]={0xff,0xff,0xff,0xff,0xff,0xff}; //全网广播ARP请求
    if(3 != argc){
            printf("Usage: %s netdevName dstIP\n",argv[0]);
            exit(1);
    }

    if(0>(skfd=socket(PF_PACKET,SOCK_RAW,htons(ETH_P_ALL)))){
            perror("Create Error");
            exit(1);
    }

    bzero(&toaddr,sizeof(toaddr));
    bzero(&ifr,sizeof(ifr));
    strcpy(ifr.ifr_name,argv[1]);

    //获取接口索引
    if(-1 == ioctl(skfd,SIOCGIFINDEX,&ifr)){
           perror("get dev index error:");
           exit(1);
    }
    toaddr.sll_ifindex = ifr.ifr_ifindex;
    printf("interface Index:%d\n",ifr.ifr_ifindex);
    //获取接口IP地址
    if(-1 == ioctl(skfd,SIOCGIFADDR,&ifr)){
           perror("get IP addr error:");
           exit(1);
    }
    srcIP.s_addr = ((struct sockaddr_in*)&(ifr.ifr_addr))->sin_addr.s_addr;
    char IP2[64];
    inet_ntop(AF_INET, (struct in_addr *)&srcIP, IP2, 64);
    printf("IP addr:%s\n",IP2);

    //获取接口的MAC地址
    if(-1 == ioctl(skfd,SIOCGIFHWADDR,&ifr)){
           perror("get dev MAC addr error:");
           exit(1);
    }

    memcpy(src_mac,ifr.ifr_hwaddr.sa_data,ETH_ALEN);
    printf("MAC :%02X-%02X-%02X-%02X-%02X-%02X\n",src_mac[0],src_mac[1],src_mac[2],src_mac[3],src_mac[4],src_mac[5]);


    //开始填充，构造以太头部
    eth=(struct ether_header*)buf;
    memcpy(eth->ether_dhost,dst_mac,ETH_ALEN);
    memcpy(eth->ether_shost,src_mac,ETH_ALEN);
    eth->ether_type = htons(ETHERTYPE_ARP);

    //手动开始填充用ARP报文首部
    arp=(struct arphdr*)(buf+sizeof(struct ether_header));
    arp->arp_hrd = htons(ARPHRD_ETHER); //硬件类型为以太
    arp->arp_pro = htons(ETHERTYPE_IP); //硬件类型为以太

    //硬件地址长度和IPV4地址长度分别是6字节和4字节
    arp->arp_hln = ETH_ALEN;
    arp->arp_pln = 4;

    //操作码，这里我们发送ARP请求
    arp->arp_op = htons(ARPOP_REQUEST);

    //填充发送端的MAC和IP地址
    memcpy(arp->arp_sha,src_mac,ETH_ALEN);
    memcpy(arp->arp_spa,&srcIP,4);

    //填充目的端的IP地址，MAC地址不用管
    inet_pton(AF_INET,argv[2],&targetIP);
    memcpy(arp->arp_tpa,&targetIP,4);

    toaddr.sll_family = PF_PACKET;
    n=sendto(skfd,buf,BUFLEN,0,(struct sockaddr*)&toaddr,sizeof(toaddr));

    close(skfd);
    return 0;
}